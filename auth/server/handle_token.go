package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	authstream "hkchat/proto/datastream/auth"
	tokenverify "hkchat/proto/tokenverify"

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
		server.logger.Print("invalid access token (check token)")
		return
	}

	user := userInfo{
		Username: claims.Username,
		UserId:   claims.UserId,
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
	exists, err := server.databaseClient.FindRefreshToken(context.Background(), &authstream.UserRefreshTokenRequest{
		RefreshToken: "refresh:" + refreshCookie.Value,
	})
	if err != nil {
		server.logger.Print("refresh redis error : ", err)
		http.Error(responseWriter, "refresh redis error", http.StatusInternalServerError)
		return
	}
	server.logger.Print("get refresh token from redis")

	if !exists.Status {
		responseWriter.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(responseWriter).Encode(authMessage{
			Status: "unauthorized",
		})
		server.logger.Print("no refresh token at redis")
		return
	}

	accessToken, err := server.generateToken(authUserRequest{Username: refreshClaims.Username, UserId: refreshClaims.UserId}, "access")
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

func (server *AuthServer) VerifyToken(ctx context.Context, request *tokenverify.VerifyTokenRequest) (*tokenverify.VerifyTokenResponse, error) {
	accessToken, ok := server.parseAccessRequestToken(request.Token)

	if !ok {
		server.logger.Print("parse auth request error (check token)")
		return nil, errors.New("parse auth request error")
	}

	var claims Claims
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &claims, checkAccessTokenMethod)
	if err != nil || !parsedAccessToken.Valid {
		server.logger.Print("invalid access token (check token)")
		return nil, errors.New("invalid access token (check token)")
	}

	return &tokenverify.VerifyTokenResponse{
		Username: claims.Username,
		UserId:   int64(claims.UserId),
	}, nil
}
