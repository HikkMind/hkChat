package server

import (
	"context"
	"errors"
	"strings"
	"time"

	authstream "hkchat/proto/datastream/auth"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId   uint
	Username string
	jwt.RegisteredClaims
}

var (
	secretKey        []byte
	refreshSecretKey []byte
)

const (
	accessTTL  time.Duration = 10 * time.Minute
	refreshTTL time.Duration = 7 * 24 * time.Hour
	// refreshTTL time.Duration = 20 * time.Second
)

func (server *AuthServer) generateToken(currentUser authUserRequest, tokenType string) (string, error) {

	var tokenKey []byte
	var tokenTTL time.Duration
	if tokenType == "access" {
		tokenKey = secretKey
		tokenTTL = accessTTL
	} else if tokenType == "refresh" {
		tokenKey = refreshSecretKey
		tokenTTL = refreshTTL
	} else {
		return "", errors.New("wrong type")
	}

	expirationTime := time.Now().Add(tokenTTL)
	claims := &Claims{
		Username: currentUser.Username,
		UserId:   currentUser.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(tokenKey)
	if err != nil {
		return "", err
	}

	server.logger.Print("genereted new " + tokenType + " token")

	return tokenString, nil
}

func (server *AuthServer) getUserInfo(authUser authUserRequest) (bool, int) {
	authResult, err := server.databaseClient.VerifyUserPassword(context.Background(), &authstream.UserDataRequest{
		Username: authUser.Username,
		Password: authUser.Password,
	})
	if err != nil {
		server.logger.Print("request error : ", err)
		return false, -1
	}
	if !authResult.Status {
		server.logger.Print("wrong login or password")
	}
	return authResult.Status, int(authResult.UserID)
}

func (server *AuthServer) generteTokenLogin(authUser authUserRequest) (string, string) {
	accessToken, err := server.generateToken(authUser, "access")
	if len(accessToken) == 0 || err != nil {
		server.logger.Print("failed generate access token : ", err)
		return "", ""
	}

	refreshToken, err := server.generateToken(authUser, "refresh")
	if len(refreshToken) == 0 || err != nil {
		server.logger.Print("failed generate refresh token : ", err)
		return "", ""
	}
	return accessToken, refreshToken
}

func (server *AuthServer) parseAccessRequestToken(requestToken string) (string, bool) {
	requestToken = strings.TrimSpace(requestToken)
	if len(requestToken) < 7 || !strings.HasPrefix(requestToken, "Bearer ") {
		server.logger.Printf("invalid token: prefix=%q, expected 'Bearer '", requestToken[:min(len(requestToken), 7)])
		return "", false
	}

	token := requestToken[7:]
	return token, true
}

func checkRefreshTokenMethod(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, errors.New("unexpected signing method")
	}
	return refreshSecretKey, nil
}

func checkAccessTokenMethod(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, errors.New("unexpected signing method")
	}
	return secretKey, nil
}
