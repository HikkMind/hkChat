package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/hikkmind/hkchat/tables"
	"github.com/lpernett/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthServer struct {
	database   *gorm.DB
	tokenUser  map[string]userInfo
	tokenMutex sync.RWMutex
	logger     *log.Logger
}

type authMessage struct {
	Status      string `json:"status"`
	AccessToken string `json:"access_token"`
	// RefreshToken string `json:"refresh_token"`
}

type authUserRequest struct {
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
	server.databaseInit()

	http.HandleFunc("/login", server.authLogin)
	http.HandleFunc("/logout", server.authLogout)
	http.HandleFunc("/register", server.authRegister)
	http.HandleFunc("/checktoken", server.authCheckToken)
	http.HandleFunc("/verifytoken", server.verifyAccessToken)

	server.logger.Print("start server")
	err := serverAuth.ListenAndServe()
	if err != nil {
		server.logger.Fatal("failed to start server")
	}
}

func (server *AuthServer) serverVariablesInit() {
	err := godotenv.Load(".dbenv")
	if err != nil {
		log.Fatal("Error loading .env file : ", err)
	}
	secretKey = []byte(os.Getenv("SECRET_KEY"))
	refreshSecretKey = []byte(os.Getenv("REFRESH_SECRET_KEY"))

	server.tokenUser = make(map[string]userInfo)
	server.logger = log.Default()
	server.logger.SetPrefix("[ AUTH ]")
}

func (server *AuthServer) databaseInit() {
	// err := godotenv.Load(".dbenv")
	// if err != nil {
	// 	log.Fatal("Error loading .env file : ", err)
	// }
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
