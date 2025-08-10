package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func (server *ChatServer) connectUser(responseWriter http.ResponseWriter, request *http.Request) {

	websocketUpgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	server.logger.Print("try connect websocket")

	connection, err := websocketUpgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		fmt.Println("failed upgrade connection : ", err)
		return
	}
	server.logger.Print("connected new websocket")

	var connMessage HandleConnectionMessage
	err = connection.ReadJSON(&connMessage)
	server.logger.Print("auth request : ", connMessage)
	// err = json.NewDecoder(request.Body).Decode(&connMessage)
	if err != nil || connMessage.Intent != "auth" {
		// http.Error(responseWriter, "failed ", http.StatusUnauthorized)
		connection.WriteJSON(HandleConnectionMessage{
			Intent: "auth",
			Status: "unauthorized",
		})
		connection.Close()
		server.logger.Print("wrong auth request")
		return
	}

	var currentUser *userInfo = server.checkAuthToken(connMessage.Token)

	if currentUser == nil {
		// http.Error(responseWriter, "unauthorized token", http.StatusUnauthorized)
		connection.WriteJSON(HandleConnectionMessage{
			Intent: "auth",
			Status: "unauthorized",
		})
		connection.Close()
		server.logger.Print("unauthorized user")
		return
	}
	currentUser.Token = connMessage.Token

	connection.WriteJSON(HandleConnectionMessage{
		Intent: "auth",
		Status: "ok",
		Token:  connMessage.Token,
	})

	server.logger.Print("handle new user : ", currentUser.Username)
	go server.handleUserConnection(connection, currentUser)
	// websocketConnection = connection
}
