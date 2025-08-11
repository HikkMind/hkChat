package server

import (
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

	websocketConnection, err := websocketUpgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		server.logger.Print("failed upgrade websocket : ", err)
		return
	}
	server.logger.Print("connected new websocket")

	currentUser := server.verifyUserToken(websocketConnection)
	if currentUser == nil {
		return
	}

	server.logger.Print("handle new user : ", currentUser.Username)
	go server.handleUserConnection(websocketConnection, currentUser)
}

func (server *ChatServer) verifyUserToken(websocketConnection *websocket.Conn) *userInfo {
	var connMessage HandleConnectionMessage
	err := websocketConnection.ReadJSON(&connMessage)
	server.logger.Print("auth request : ", connMessage)
	if err != nil || connMessage.Intent != "auth" {
		websocketConnection.WriteJSON(HandleConnectionMessage{
			Intent: "auth",
			Status: "unauthorized",
		})
		websocketConnection.Close()
		server.logger.Print("wrong auth request")
		return nil
	}

	var currentUser *userInfo = server.checkAuthToken(connMessage.Token)

	if currentUser == nil {
		websocketConnection.WriteJSON(HandleConnectionMessage{
			Intent: "auth",
			Status: "unauthorized",
		})
		websocketConnection.Close()
		server.logger.Print("unauthorized user")
		return nil
	}
	currentUser.Token = connMessage.Token

	websocketConnection.WriteJSON(HandleConnectionMessage{
		Intent: "auth",
		Status: "ok",
		Token:  connMessage.Token,
	})

	return currentUser
}
