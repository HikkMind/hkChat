package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hikkmind/hkchat/server/tables"
	"github.com/hikkmind/hkchat/structs"
	"github.com/lpernett/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	database *gorm.DB
)

func main() {
	serverHTTP := &http.Server{
		Addr: ":8081",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	databaseInit()

	http.HandleFunc("/login", AuthLogin)
	http.HandleFunc("/register", AuthRegister)
	err := serverHTTP.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start server")
	}

	fmt.Println("ended")

}

func AuthLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}

	result := database.Where("username = ? AND password = ?", authUser.Username, authUser.Password).First(&tables.User{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("login error")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	} else if result.Error != nil {
		fmt.Println("request error : ", result.Error.Error())
		return
	}

	fmt.Println("user " + authUser.Username + " logged in")

	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write([]byte(`{"status":"ok"}`))

}

func AuthRegister(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}

	if len(authUser.Username) == 0 || len(authUser.Password) == 0 {
		responseWriter.WriteHeader(http.StatusConflict)
	}

	result := database.Create(&tables.User{Username: authUser.Username, Password: authUser.Password})
	if result.Error != nil {
		fmt.Println("duplicate error")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	}

	fmt.Println("new user : ", authUser.Username)
	responseWriter.WriteHeader(http.StatusCreated)

}

func databaseInit() {
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

	database = db

	fmt.Println("connected to database")
}
