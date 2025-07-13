package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hikkmind/hkchat/server/tables"
	"github.com/hikkmind/hkchat/structs"
	"gorm.io/gorm"
)

func MessageHandler(responseWriter http.ResponseWriter, request *http.Request) {

	connection, err := upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		fmt.Println("failed upgrade connection : ", err)
		return
	}
	websocketList[connection] = make(chan []byte, startMessagesCount)

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

func AuthLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}

	// result := database.First(&tables.User{Username: authUser.Username, Password: authUser.Password})
	result := database.Where("username = ? AND password = ?", authUser.Username, authUser.Password).First(&tables.User{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("login error")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	} else if result.Error != nil {
		fmt.Println("request error : ", result.Error.Error())
		return
	}

	// username := authUser.Username
	// password := authUser.Password

	// if actPass, ok := usernameList[username]; !ok {
	// 	statusAnswer(responseWriter, "unregistered user", http.StatusConflict)
	// 	return
	// } else if password != actPass {
	// 	statusAnswer(responseWriter, "wrong password", http.StatusConflict)
	// 	return
	// }

	fmt.Println("user " + authUser.Username + " logged in")

	responseWriter.WriteHeader(http.StatusOK)

}

func AuthRegister(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}
	// username := authUser.Username

	result := database.Create(&tables.User{Username: authUser.Username, Password: authUser.Password})
	if result.Error != nil {
		fmt.Println("duplicate error")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	}

	fmt.Println("new user : ", authUser.Username)
	responseWriter.WriteHeader(http.StatusCreated)

	// if _, ok := usernameList[username]; ok {
	// 	fmt.Printf("user %s already exists\n", username)
	// 	statusAnswer(responseWriter, "user "+username+" already exists", http.StatusConflict)
	// 	return
	// }

	// conn := request.Context().Value("connection").(net.Conn)
	// if actUsername, ok := usersConList[conn]; ok {
	// 	mess := structs.MessageStatus{Message: "changed username from " + actUsername + " to " + username}
	// 	responseWriter.Header().Set("Content-Type", "application/json")
	// 	responseWriter.WriteHeader(http.StatusCreated)
	// 	json.NewEncoder(responseWriter).Encode(mess)
	// 	usersConList[conn] = username
	// 	delete(usernameList, actUsername)
	// 	usernameList[username] = authUser.Password
	// 	return
	// }

	// usernameList[username] = authUser.Password
	// usersConList[conn] = username
	// fmt.Println("new user : ", username)

}
