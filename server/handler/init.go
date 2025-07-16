package handler

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/server/tables"
	"github.com/lpernett/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	mainServer Server
)

type Server struct {
	UsersConList       map[net.Conn]string
	UsernameList       map[string]string
	WebsocketList      map[*websocket.Conn]chan []byte
	Database           *gorm.DB
	StartMessagesCount int

	Upgrader websocket.Upgrader
}

// var (
// 	usersConList  map[net.Conn]string
// 	usernameList  map[string]string
// 	websocketList map[*websocket.Conn]chan []byte
// 	// websocketList      map[string]*websocket.Conn
// 	// usersWebsocketList map[*websocket.Conn]string

// 	database           *gorm.DB
// 	startMessagesCount int

// 	upgrader websocket.Upgrader = websocket.Upgrader{
// 		CheckOrigin: func(r *http.Request) bool {
// 			return true
// 		},
// 	}
// )

func (server *Server) Start() error {
	serverHTTP := &http.Server{
		Addr: ":8080",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	server.databaseInit()
	server.handlerInit()

	http.HandleFunc("/messager", MessageHandler)
	http.HandleFunc("/login", AuthLogin)
	http.HandleFunc("/register", AuthRegister)
	fmt.Println("server started")
	// err := http.ListenAndServe("localhost:8080", nil)
	return serverHTTP.ListenAndServe()
}

func (server *Server) handlerInit() {
	server.UsernameList = make(map[string]string)
	server.UsersConList = make(map[net.Conn]string)
	// websocketList = make(map[string]*websocket.Conn)
	server.WebsocketList = make(map[*websocket.Conn]chan []byte)

	server.Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (server *Server) databaseInit() {
	err := godotenv.Load(".dbenv")
	if err != nil {
		log.Fatal("Error loading .env file : ", err)
	}
	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}

	server.Database = db
	server.StartMessagesCount, err = strconv.Atoi(os.Getenv("START_MESSAGE_COUNT"))
	if err != nil {
		fmt.Println("database init error : ", err)
		return
	}

	fmt.Println("connected to database")
}

func StartServer() error {
	mainServer = Server{}
	return mainServer.Start()
}
