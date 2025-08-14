package server

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (server *AuthServer) authCheckToken(responseWriter http.ResponseWriter, request *http.Request) {

	responseWriter.Header().Set("Content-Type", "application/json")
	accessTokenRequestString := request.Header.Get("Authorization")

	accessToken, ok := server.parseAccessRequestToken(accessTokenRequestString)

	if !ok {
		http.Error(responseWriter, "parse auth request error", http.StatusBadRequest)
		server.logger.Print("parse auth request error (check token)")
		return
	}

	var claims Claims
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &claims, checkAccessTokenMethod)
	if err != nil || !parsedAccessToken.Valid {
		http.Error(responseWriter, "invalid auth token", http.StatusUnauthorized)
		server.redisDatabase.Del(server.redisContext, accessToken)
		server.logger.Print("invalid access token (check token)")
		return
	}

	user := userInfo{
		Username: claims.Username,
		UserId:   claims.UserID,
	}

	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(user)

}

func (server *AuthServer) verifyAccessToken(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	tokenRequestString := request.Header.Get("Authorization")

	authToken, ok := server.parseAccessRequestToken(tokenRequestString)
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

	server.logger.Print("update access token...")
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

	server.logger.Print("try get " + "refresh:" + refreshCookie.Value)
	exists, err := server.redisDatabase.Exists(server.redisContext, "refresh:"+refreshCookie.Value).Result()
	if err != nil {
		server.logger.Print("refresh redis error : ", err)
		http.Error(responseWriter, "refresh redis error", http.StatusInternalServerError)
		return
	}
	server.logger.Print("get refresh token from redis")

	if exists == 0 {
		// http.Error(responseWriter, "refresh token expired or revoked", http.StatusUnauthorized)
		responseWriter.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(responseWriter).Encode(authMessage{
			Status: "unauthorized",
		})
		server.logger.Print("no refresh token at redis")
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
	server.logger.Print("access token updated")

}
