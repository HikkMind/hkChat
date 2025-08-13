package server

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (server *AuthServer) authCheckToken(responseWriter http.ResponseWriter, request *http.Request) {

	responseWriter.Header().Set("Content-Type", "application/json")
	tokenRequestString := request.Header.Get("Authorization")

	authToken, ok := server.parseRequestToken(tokenRequestString)

	if !ok {
		http.Error(responseWriter, "parse auth request error", http.StatusBadRequest)
		server.logger.Print("parse auth request error (check token)")
		return
	}

	var claims Claims
	parsedToken, err := jwt.ParseWithClaims(authToken, &claims, checkAccessTokenMethod)
	if err != nil || !parsedToken.Valid {
		http.Error(responseWriter, "invalid auth token", http.StatusUnauthorized)
		server.logger.Print("invalid auth token (check token)")
		return
	}

	server.tokenMutex.RLock()
	if user, ok := server.tokenUser[authToken]; ok {
		err := json.NewEncoder(responseWriter).Encode(user)
		server.logger.Print("accept token of user : ", user.Username)
		if err != nil {
			http.Error(responseWriter, "internal error", http.StatusInternalServerError)
			server.logger.Print("send response error : ", err)
		}
		server.tokenMutex.RUnlock()
		return
	}
	server.tokenMutex.RUnlock()

	responseWriter.WriteHeader(http.StatusUnauthorized)
	server.logger.Print("user unauthorized")

}

func (server *AuthServer) verifyAccessToken(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	tokenRequestString := request.Header.Get("Authorization")

	authToken, ok := server.parseRequestToken(tokenRequestString)
	if !ok {
		http.Error(responseWriter, "parse auth access error", http.StatusBadRequest)
		server.logger.Print("parse auth access error (verify token)")
		return
	}
	var claims Claims
	parsedToken, err := jwt.ParseWithClaims(authToken, &claims, checkAccessTokenMethod)
	if err == nil && parsedToken.Valid {
		responseWriter.WriteHeader(http.StatusOK)
		json.NewEncoder(responseWriter).Encode(authMessage{Status: "ok"})
		return
	}

	refreshCookie, err := request.Cookie("refresh_token")
	if err != nil {
		http.Error(responseWriter, "parse auth cookie error", http.StatusUnauthorized)
		server.logger.Print("parse auth refresh error (verify token): ", err)
		return
	}
	var refreshClaims Claims
	parsedToken, err = jwt.ParseWithClaims(refreshCookie.Value, &refreshClaims, checkRefreshTokenMethod)

	if err != nil || !parsedToken.Valid {
		http.Error(responseWriter, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, err := server.generateToken(authUserRequest{Username: refreshClaims.Username}, "access")
	if err != nil {
		http.Error(responseWriter, "failed to create new access token", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(authMessage{
		Status:      "refresh",
		AccessToken: accessToken,
	})

}
