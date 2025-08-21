package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	tokenverify "github.com/hikkmind/hkchat/proto/tokenverify"
	"github.com/hikkmind/hkchat/tables"
	"github.com/lpernett/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthServer struct {
	database           *gorm.DB
	redisDatabase      *redis.Client
	redisContext       context.Context
	redisContextCancel context.CancelFunc

	// tokenUser  map[string]userInfo
	// tokenMutex sync.RWMutex
	logger *log.Logger
	tokenverify.UnimplementedAuthServiceServer
}

type authMessage struct {
	Status      string `json:"status"`
	AccessToken string `json:"access_token"`
	// RefreshToken string `json:"refresh_token"`
}

type authUserRequest struct {
	UserId   uint   `json:"-"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type userInfo struct {
	Username string `json:"username"`
	UserId   uint   `json:"user_id"`
}

func (server *AuthServer) StartServer() {
	serverAuth := &http.Server{
		Addr: ":8081",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	server.serverVariablesInit()

	// grpcConnectionContext, grpcConnectionSignal := context.WithCancel(context.Background())
	go server.startGrpcServer()
	// <-grpcConnectionContext.Done()

	server.redisInit()
	server.databaseInit()

	http.HandleFunc("/login", server.authLogin)
	http.HandleFunc("/logout", server.authLogout)
	http.HandleFunc("/register", server.authRegister)
	// http.HandleFunc("/checktoken", server.authCheckToken)
	http.HandleFunc("/verifytoken", server.verifyAccessToken)

	server.logger.Print("start server")
	err := serverAuth.ListenAndServe()
	if err != nil {
		server.logger.Fatal("failed to start server")
	}
}

func (server *AuthServer) serverVariablesInit() {

	server.logger = log.Default()
	server.logger.SetPrefix("[ AUTH ]")

	err := godotenv.Load(".dbenv")
	if err != nil {
		server.logger.Fatal("Error loading .env file : ", err)
	}
	secretKey = []byte(os.Getenv("SECRET_KEY"))
	refreshSecretKey = []byte(os.Getenv("REFRESH_SECRET_KEY"))

	usernameMinLength, err = strconv.Atoi(os.Getenv("USERNAME_LENGTH"))
	passwordMinLength, err = strconv.Atoi(os.Getenv("PASSWORD_LENGTH"))
	if err != nil {
		server.logger.Fatal("error get user settings from environment : ", err)
	}

	server.redisContext, server.redisContextCancel = context.WithCancel(context.Background())
}

func (server *AuthServer) databaseInit() {
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	server.database = db
	server.logger.Print("connected to database")
}

func (server *AuthServer) redisInit() {
	server.redisDatabase = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := server.redisDatabase.Ping(context.Background()).Result()
	if err != nil {
		server.logger.Fatal("failed connect redis: ", err)
		return
	}

	server.logger.Print("connected to redis")
}
