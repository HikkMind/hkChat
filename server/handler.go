package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/structs"
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

func handlerInit() {
	usernameList = make(map[string]string)
	usersConList = make(map[net.Conn]string)
	// websocketList = make(map[string]*websocket.Conn)
	websocketList = make(map[*websocket.Conn]struct{})
}

func userHandler(connection *websocket.Conn) {
	defer connection.Close()
	defer delete(websocketList, connection)
	for {
		messageType, msg, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		// fmt.Println(string(msg))
		// fmt.Println("got from ", username, " : ", string(message))

		for conn := range websocketList {
			if connection == conn {
				continue
			}
			// data, _ := json.Marshal(structs.Message{Sender: username, Message: string(msg)})
			err = conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func messageHandler(responseWriter http.ResponseWriter, request *http.Request) {

	connection, err := upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		fmt.Println("failed upgrade connection : ", err)
		return
	}
	websocketList[connection] = struct{}{}

	// username := usersConList[request.Context().Value("connection").(net.Conn)]
	go userHandler(connection)

	// var message structs.Message
	// err := json.NewDecoder(request.Body).Decode(&message)
	// if err != nil {
	// 	fmt.Println("error json : ", err)
	// 	return
	// }
	// if len(message.Message) > 0 {
	// 	fmt.Println(username, " : ", message.Message)
	// }
}

func authLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}
	username := authUser.Username
	password := authUser.Password

	if actPass, ok := usernameList[username]; !ok {
		statusAnswer(responseWriter, "unregistered user", http.StatusConflict)
		return
	} else if password != actPass {
		statusAnswer(responseWriter, "wrong password", http.StatusConflict)
		return
	}

	fmt.Println("user " + username + " logged in")

	responseWriter.WriteHeader(http.StatusOK)

}

func authRegister(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}
	username := authUser.Username

	if _, ok := usernameList[username]; ok {
		fmt.Printf("user %s already exists\n", username)
		statusAnswer(responseWriter, "user "+username+" already exists", http.StatusConflict)
		return
	}

	conn := request.Context().Value("connection").(net.Conn)
	if actUsername, ok := usersConList[conn]; ok {
		mess := structs.MessageStatus{Message: "changed username from " + actUsername + " to " + username}
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusCreated)
		json.NewEncoder(responseWriter).Encode(mess)
		usersConList[conn] = username
		delete(usernameList, actUsername)
		usernameList[username] = authUser.Password
		return
	}

	usernameList[username] = authUser.Password
	usersConList[conn] = username
	fmt.Println("new user : ", username)
	responseWriter.WriteHeader(http.StatusCreated)

}

func handleConnectionReader(connection net.Conn, username string) {

	clientMessage := make([]byte, 1024)
	connection.Write([]byte("You was connected with username : " + username))
	defer connection.Close()

	for {
		n, err := connection.Read(clientMessage)
		if err != nil {
			fmt.Printf("client %s disconnected\n", username)
			break
		}
		fmt.Printf("client %s sent message: %s\n", username, string(clientMessage[:n]))
	}
}

func statusAnswer(responseWriter http.ResponseWriter, message string, code int) {
	mess := structs.MessageStatus{Message: message}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	json.NewEncoder(responseWriter).Encode(mess)
}

// func handleSendMessage(connection net.Conn, message []byte) {
// 	connection.Write(message)
// }
