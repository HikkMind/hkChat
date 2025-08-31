package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	authstream "hkchat/proto/datastream/auth"
	tokenverify "hkchat/proto/tokenverify"

	"github.com/lpernett/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServer struct {
	// database           *gorm.DB
	redisDatabase      *redis.Client
	redisContext       context.Context
	redisContextCancel context.CancelFunc

	// tokenUser  map[string]userInfo
	// tokenMutex sync.RWMutex
	logger         *log.Logger
	databaseClient authstream.UserDataServiceClient
	tokenverify.UnimplementedAuthServiceServer
	// authstream.UserDataServiceClient
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
	go server.startGrpcServer()

	server.redisInit()
	server.databaseInit()

	http.HandleFunc("/login", server.authLogin)
	http.HandleFunc("/logout", server.authLogout)
	http.HandleFunc("/register", server.authRegister)
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
	// dsn := os.Getenv("DB_CONFIG")
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	log.Fatal("failed to connect:", err)
	// }
	// if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
	// 	log.Fatal("migration failed:", err)
	// }
	// server.database = db
	// server.logger.Print("connected to database")
	tokenConnection, err := grpc.NewClient("database:6002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.logger.Print("failed check auth token : ", err)
		return
	}
	server.databaseClient = authstream.NewUserDataServiceClient(tokenConnection)
	server.logger.Print("connected to grpc server")
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
