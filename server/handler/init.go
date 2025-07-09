package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	usersConList  map[net.Conn]string
	usernameList  map[string]string
	websocketList map[*websocket.Conn]struct{}
	// websocketList      map[string]*websocket.Conn
	// usersWebsocketList map[*websocket.Conn]string

	upgrader websocket.Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func StartServer() error {
	mainServer := &http.Server{
		Addr: ":8080",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	handlerInit()

	http.HandleFunc("/messager", MessageHandler)
	http.HandleFunc("/login", AuthLogin)
	http.HandleFunc("/register", AuthRegister)
	fmt.Println("server started")
	// err := http.ListenAndServe("localhost:8080", nil)
	return mainServer.ListenAndServe()
}

func handlerInit() {
	usernameList = make(map[string]string)
	usersConList = make(map[net.Conn]string)
	// websocketList = make(map[string]*websocket.Conn)
	websocketList = make(map[*websocket.Conn]struct{})
}
