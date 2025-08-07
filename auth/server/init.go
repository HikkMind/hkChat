package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/hikkmind/hkchat/server/tables"
	"github.com/lpernett/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthServer struct {
	database   *gorm.DB
	tokenUser  map[string]userInfo
	tokenMutex sync.RWMutex
}

type authMessage struct {
	Status string `json:"status"`
	Token  string `json:"token"`
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

	server.tokenUser = make(map[string]userInfo)
	server.databaseInit()

	http.HandleFunc("/login", server.authLogin)
	http.HandleFunc("/logout", server.authLogout)
	http.HandleFunc("/register", server.authRegister)
	http.HandleFunc("/checktoken", server.authCheckToken)
	err := serverAuth.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start server")
	}
}

func (server *AuthServer) databaseInit() {
	err := godotenv.Load(".dbenv")
	if err != nil {
		log.Fatal("Error loading .env file : ", err)
	}
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	server.database = db

	fmt.Println("connected to database")
}
