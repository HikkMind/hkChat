package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	authstream "hkchat/proto/datastream/auth"
)

var (
	usernameMinLength int
	passwordMinLength int
)

func (server *AuthServer) authLogin(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser authUserRequest
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		server.logger.Print("failed decode login request : ", err)
		return
	}

	ok, userID := server.getUserInfo(authUser)
	if !ok {
		http.Error(responseWriter, "wrong login or password", http.StatusUnauthorized)
		return
	}

	authUser.UserId = uint(userID)

	accessToken, refreshToken := server.generteTokenLogin(authUserRequest{Username: authUser.Username, UserId: authUser.UserId})
	if len(accessToken) == 0 || len(refreshToken) == 0 {
		http.Error(responseWriter, "failed generate access/refresh token", http.StatusInternalServerError)
	}

	server.logger.Print("user logged in : ", authUser.Username)
	// server.redisDatabase.Set(server.redisContext, "refresh:"+refreshToken, "", refreshTTL)
	server.databaseClient.SetRefreshToken(context.Background(), &authstream.UserRefreshTokenRequest{
		RefreshToken: "refresh:" + refreshToken,
	})

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

	refreshCookie, _ := request.Cookie("refresh_token")

	// server.redisDatabase.Del(server.redisContext, "refresh:"+refreshCookie.Value).Err()
	server.databaseClient.UnsetRefreshToken(context.Background(), &authstream.UserRefreshTokenRequest{
		RefreshToken: "refresh:" + refreshCookie.Value,
	})

	responseWriter.WriteHeader(http.StatusOK)
}

func (server *AuthServer) authRegister(responseWriter http.ResponseWriter, request *http.Request) {
	var authUser authUserRequest
	err := json.NewDecoder(request.Body).Decode(&authUser)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		server.logger.Print("failed decode register request : ", err)
		return
	}

	if len(authUser.Username) < usernameMinLength || len(authUser.Password) < passwordMinLength {
		responseWriter.WriteHeader(http.StatusUnauthorized)
		server.logger.Print("(REGISTER) empty login or password")
	}

	// result := server.database.Create(&tables.User{Username: authUser.Username, Password: authUser.Password})
	_, err = server.databaseClient.RegisterNewUser(context.Background(), &authstream.UserDataRequest{
		Username: authUser.Username,
		Password: authUser.Password,
	})
	if err != nil {
		server.logger.Print("failed register new user : ", err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	server.logger.Print("register new user : ", authUser.Username)
	responseWriter.WriteHeader(http.StatusCreated)

}
