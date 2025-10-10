package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	authstream "hkchat/proto/datastream/auth"
	tokenverify "hkchat/proto/tokenverify"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServer struct {
	// database           *gorm.DB

	// tokenUser  map[string]userInfo
	// tokenMutex sync.RWMutex
	logger         *log.Logger
	databaseClient authstream.UserDataServiceClient
	tokenverify.UnimplementedAuthServiceServer

	serverPort string
	grpcPort   string
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
	server.serverVariablesInit()

	serverAuth := &http.Server{
		Addr: server.serverPort,
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	go server.startGrpcServer()

	// server.redisInit()
	server.databaseInit()

	http.HandleFunc("/login", server.authLogin)
	http.HandleFunc("/logout", server.authLogout)
	http.HandleFunc("/register", server.authRegister)
	http.HandleFunc("/verifytoken", server.verifyAccessToken)

	server.logger.Print("start server on port ", server.serverPort)
	err := serverAuth.ListenAndServe()
	if err != nil {
		server.logger.Fatal("failed to start server")
	}
}

func (server *AuthServer) serverVariablesInit() {

	server.serverPort = ":" + os.Getenv("AUTH_PORT")
	server.grpcPort = ":" + os.Getenv("AUTH_GRPC_PORT")

	server.logger = log.Default()
	server.logger.SetPrefix("[ AUTH ]")
	secretKey = []byte(os.Getenv("SECRET_KEY"))
	refreshSecretKey = []byte(os.Getenv("REFRESH_SECRET_KEY"))

	// usernameMinLength, err := strconv.Atoi(os.Getenv("USERNAME_LENGTH"))
	// passwordMinLength, err = strconv.Atoi(os.Getenv("PASSWORD_LENGTH"))
	// if err != nil {
	// 	server.logger.Fatal("error get user settings from environment : ", err)
	// }

	// server.redisContext, server.redisContextCancel = context.WithCancel(context.Background())
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

	datagatePort := ":" + os.Getenv("DATAGATE_GRPC_PORT")

	dataConnection, err := grpc.NewClient("datagate"+datagatePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.logger.Print("failed connect to datagate : ", err)
		return
	}
	server.databaseClient = authstream.NewUserDataServiceClient(dataConnection)
	server.logger.Print("connected to grpc server")
}
