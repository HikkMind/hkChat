package server

import (
	"context"
	authstream "hkchat/proto/datastream/auth"
	chatstream "hkchat/proto/datastream/chat"
	"log"
	"os"
	"time"

	"hkchat/tables"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseServer struct {
	logger *log.Logger
	// databaseClient authstream.UserDataServiceClient

	databaseConnection *gorm.DB
	redisConnection    *redis.Client
	redisContext       context.Context
	redisContextCancel context.CancelFunc
	authstream.UnimplementedUserDataServiceServer
	chatstream.UnimplementedChatServiceServer

	refreshTTL time.Duration
}

func NewServer() *DatabaseServer {
	server := &DatabaseServer{}

	server.logger = log.Default()
	server.logger.SetPrefix("[ DATABASE ]")

	server.postgresInit()
	server.redisInit()

	server.refreshTTL = 7 * 24 * time.Hour

	return server
}

func (server *DatabaseServer) StartServer() {
	go server.startGrpcServer()
}

func (server *DatabaseServer) postgresInit() {
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}
	server.databaseConnection = db
	server.logger.Print("connected to database")
}

func (server *DatabaseServer) redisInit() {
	server.redisConnection = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := server.redisConnection.Ping(context.Background()).Result()
	if err != nil {
		server.logger.Fatal("failed connect redis: ", err)
		return
	}

	server.logger.Print("connected to redis")
}
