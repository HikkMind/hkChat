package server

import (
	"context"
	"fmt"
	authstream "hkchat/proto/datastream/auth"
	chatstream "hkchat/proto/datastream/chat"
	"log"
	"os"
	"time"

	"hkchat/tables"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DatabaseServer struct {
	logger *log.Logger

	databaseConnection *gorm.DB
	redisConnection    *redis.Client
	redisContext       context.Context
	redisContextCancel context.CancelFunc
	authstream.UnimplementedUserDataServiceServer
	chatstream.UnimplementedChatServiceServer

	refreshTTL time.Duration

	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
}

func NewServer() *DatabaseServer {
	server := &DatabaseServer{}

	server.logger = log.Default()
	server.logger.SetPrefix("[ DATABASE ]")

	server.refreshTTL = 7 * 24 * time.Hour

	return server
}

func (server *DatabaseServer) StartServer() {

	server.logger.Print("starting server at time : ", time.Now(), "...")

	server.postgresInit()
	server.redisInit()
	server.rabbitInit()
	server.handleRabbitConnection()

	server.startGrpcServer()
}

func (server *DatabaseServer) postgresInit() {

	server.logger.Print("postgres connecting...")

	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}
	server.databaseConnection = db
	server.logger.Print("connected to postgres")
}

func (server *DatabaseServer) redisInit() {

	server.logger.Print("redis connecting...")

	server.redisConnection = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	connResult, err := server.redisConnection.Ping(context.Background()).Result()
	server.logger.Print("redis connection : ", connResult)
	if err != nil {
		server.logger.Fatal("failed connect redis: ", err)
		return
	}

	server.redisContext, server.redisContextCancel = context.WithCancel(context.Background())

	server.logger.Print("connected to redis")
}

func (server *DatabaseServer) rabbitInit() {
	server.logger.Print("rabbitmq connecting...")
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
		os.Getenv("RABBITMQ_VHOST"),
	)
	// url := os.Getenv("RABBITMQ_URL") // amqp://guest:guest@rabbitmq:5672/
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ:", err)
	}
	server.rabbitConn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("failed to open channel:", err)
	}
	server.rabbitChannel = ch

	// Объявляем exchange (должен совпадать с бекендом)
	err = ch.ExchangeDeclare("chat.events", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal("failed to declare exchange:", err)
	}

	err = ch.ExchangeDeclare("chat.dlx", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal("failed to declare DLX:", err)
	}

	// Объявляем очередь для datagate
	q, err := ch.QueueDeclare("datagate.queue", true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "chat.dlx",
		"x-dead-letter-routing-key": "datagate.dlq",
	})
	if err != nil {
		log.Fatal("failed to declare queue:", err)
	}

	_, err = ch.QueueDeclare("datagate.dlq", true, false, false, false, nil)
	if err != nil {
		log.Fatal("failed to declare DLQ:", err)
	}

	err = ch.QueueBind(q.Name, "chat.*.event", "chat.events", false, nil)
	if err != nil {
		log.Fatal("failed to bind queue:", err)
	}

	_, err = ch.QueueDeclare("datagate.dlq", true, false, false, false, nil)
	if err != nil {
		log.Fatal("failed to declare DLQ:", err)
	}

	server.logger.Print("rabbitmq connected")
}
