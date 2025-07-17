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

	connection, err := mainServer.Upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		fmt.Println("failed upgrade connection : ", err)
		return
	}
	mainServer.WebsocketList[connection] = make(chan []byte, mainServer.StartMessagesCount)

	go userHandler(connection)
}

func AuthLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser structs.AuthUser
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		fmt.Println("failed decode auth")
		return
	}

	result := mainServer.Database.Where("username = ? AND password = ?", authUser.Username, authUser.Password).First(&tables.User{})
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

	result := mainServer.Database.Create(&tables.User{Username: authUser.Username, Password: authUser.Password})
	if result.Error != nil {
		fmt.Println("duplicate error")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	}

	fmt.Println("new user : ", authUser.Username)
	responseWriter.WriteHeader(http.StatusCreated)

}
