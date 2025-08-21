package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hikkmind/hkchat/connection/chat"
	tokenverify "github.com/hikkmind/hkchat/proto/tokenverify"
	"github.com/hikkmind/hkchat/tables"
	"github.com/lpernett/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ChatServer struct {
	database     *gorm.DB
	chatList     map[uint]chan chat.ControlMessage
	chatListName map[uint]string
	logger       *log.Logger

	authTokenClient tokenverify.AuthServiceClient
}

type HandleConnectionMessage struct {
	Intent string `json:"intent"`
	Status string `json:"status"`
	ChatId int    `json:"chat_id"`
	Token  string `json:"token"`
	Text   string `json:"text"`
}

type userInfo struct {
	Username string `json:"username"`
	UserId   uint   `json:"user_id"`
	Token    string `json:"token"`
}

func (server *ChatServer) StartServer() {
	serverHTTP := &http.Server{
		Addr: ":8080",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	http.HandleFunc("/chatlist", server.connectUser)
	server.logger = log.Default()
	server.logger.SetPrefix("[ CONNECTION ]")

	server.databaseInit()
	server.grpcInit()
	server.loadChats()

	serverHTTP.ListenAndServe()

}

func (server *ChatServer) databaseInit() {
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

func (server *ChatServer) grpcInit() {
	tokenConnection, err := grpc.NewClient("auth:6001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.logger.Print("failed check auth token : ", err)
		return
	}
	// defer tokenConnection.Close()

	server.authTokenClient = tokenverify.NewAuthServiceClient(tokenConnection)
	server.logger.Print("connected to grpc server")
}

func (server *ChatServer) loadChats() {
	allChats := make([]tables.Chat, 0)
	server.database.Table("chats").Find(&allChats)

	server.chatList = make(map[uint]chan chat.ControlMessage)
	server.chatListName = make(map[uint]string)

	for _, currentChat := range allChats {
		chatChannel := make(chan chat.ControlMessage)

		server.chatList[currentChat.ID] = chatChannel
		server.chatListName[currentChat.ID] = currentChat.Name

		go chat.HandleChat(chatChannel, currentChat.ID, server.database)
	}
}
