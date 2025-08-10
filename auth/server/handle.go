package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/hikkmind/hkchat/tables"
	"gorm.io/gorm"
)

func (server *AuthServer) authLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser authUserRequest
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		server.logger.Print("failed decode login request : ", err)
		return
	}

	var user tables.User
	result := server.database.Where("username = ? AND password = ?", authUser.Username, authUser.Password).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		server.logger.Print("wrong login or password")
		responseWriter.WriteHeader(http.StatusConflict)
		return
	} else if result.Error != nil {
		server.logger.Print("request error : ", result.Error.Error())
		return
	}

	authToken := strconv.Itoa(int(user.ID))
	server.tokenUser[authToken] = userInfo{Username: user.Username, UserId: user.ID}

	server.logger.Print("user logged in : ", authUser.Username)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)

	json.NewEncoder(responseWriter).Encode(authMessage{
		Status: "ok",
		Token:  authToken,
	})

}

func (server *AuthServer) authLogout(responseWriter http.ResponseWriter, request *http.Request) {

	var logoutMessage authMessage
	err := json.NewDecoder(request.Body).Decode(&logoutMessage)
	if err != nil {
		http.Error(responseWriter, "parse_json_error", http.StatusBadRequest)
	}

	server.tokenMutex.RLock()
	if _, ok := server.tokenUser[logoutMessage.Token]; !ok {
		http.Error(responseWriter, "wrong_token", http.StatusBadRequest)
	}

	server.tokenMutex.RUnlock()
	server.tokenMutex.Lock()
	delete(server.tokenUser, logoutMessage.Token)
	server.tokenMutex.Unlock()

	responseWriter.WriteHeader(http.StatusOK)
}

func (server *AuthServer) authRegister(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser authUserRequest
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		server.logger.Print("failed decode register request : ", err)
		return
	}

	if len(authUser.Username) == 0 || len(authUser.Password) == 0 {
		responseWriter.WriteHeader(http.StatusConflict)
		server.logger.Print("(REGISTER) empty login or password")
	}

	result := server.database.Create(&tables.User{Username: authUser.Username, Password: authUser.Password})
	if result.Error != nil {
		server.logger.Print("failed register new user : ", result.Error.Error())
		responseWriter.WriteHeader(http.StatusConflict)
		return
	}

	server.logger.Print("register new user : ", authUser.Username)
	responseWriter.WriteHeader(http.StatusCreated)

}

func (server *AuthServer) authCheckToken(responseWriter http.ResponseWriter, request *http.Request) {

	responseWriter.Header().Set("Content-Type", "application/json")

	var checkAuthRequest authMessage
	err := json.NewDecoder(request.Body).Decode(&checkAuthRequest)
	server.logger.Print("get new token check request")
	if err != nil {
		http.Error(responseWriter, "parse_json_error", http.StatusBadRequest)
		server.logger.Print("parse auth message error : ", err)
		return
	}

	server.tokenMutex.RLock()
	defer server.tokenMutex.RUnlock()
	if user, ok := server.tokenUser[checkAuthRequest.Token]; ok {
		err = json.NewEncoder(responseWriter).Encode(user)
		server.logger.Print("accept token of user : ", user.Username)
		if err != nil {
			http.Error(responseWriter, "internal error", http.StatusInternalServerError)
			server.logger.Print("send response error : ", err)
		}
		return
	}

	responseWriter.WriteHeader(http.StatusUnauthorized)
	server.logger.Print("user unauthorized")

}
