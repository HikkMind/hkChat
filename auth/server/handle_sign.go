package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

	// authToken := strconv.Itoa(int(user.ID))
	accessToken, err := server.generateToken(authUser, "access")
	if len(accessToken) == 0 || err != nil {
		server.logger.Print("failed generate access token : ", err)
		http.Error(responseWriter, "failed generate access token", http.StatusInternalServerError)
		return
	}
	server.tokenUser[accessToken] = userInfo{Username: user.Username, UserId: user.ID}

	refreshToken, err := server.generateToken(authUser, "refresh")
	if len(refreshToken) == 0 || err != nil {
		server.logger.Print("failed generate refresh token : ", err)
		http.Error(responseWriter, "failed generate refresh token", http.StatusInternalServerError)
		return
	}

	server.logger.Print("user logged in : ", authUser.Username)

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(refreshTTL),
	})

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)

	json.NewEncoder(responseWriter).Encode(authMessage{
		Status:      "ok",
		AccessToken: accessToken,
	})

}

func (server *AuthServer) authLogout(responseWriter http.ResponseWriter, request *http.Request) {

	var logoutMessage authMessage
	err := json.NewDecoder(request.Body).Decode(&logoutMessage)
	if err != nil {
		http.Error(responseWriter, "parse_json_error", http.StatusBadRequest)
	}

	server.tokenMutex.RLock()
	if _, ok := server.tokenUser[logoutMessage.AccessToken]; !ok {
		http.Error(responseWriter, "wrong_token", http.StatusBadRequest)
	}

	server.tokenMutex.RUnlock()
	server.tokenMutex.Lock()
	delete(server.tokenUser, logoutMessage.AccessToken)
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
