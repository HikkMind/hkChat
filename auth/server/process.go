package server

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint
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
		UserID:   currentUser.UserId,
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

func (server *AuthServer) parseAccessRequestToken(requestToken string) (string, bool) {
	// server.logger.Printf("request token (raw): %q, length: %d", requestToken, len(requestToken))
	requestToken = strings.TrimSpace(requestToken)
	if len(requestToken) < 7 || !strings.HasPrefix(requestToken, "Bearer ") {
		server.logger.Printf("invalid token: prefix=%q, expected 'Bearer '", requestToken[:min(len(requestToken), 7)])
		return "", false
	}

	token := requestToken[7:]
	// server.logger.Printf("extracted token: %q", token)
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
