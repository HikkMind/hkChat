package server

import (
	"context"

	tokenverify "github.com/hikkmind/hkchat/proto/tokenverify"
)

type authMessage struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func (server *ChatServer) checkAuthToken(token string) *userInfo {

	authResponse, err := server.authTokenClient.VerifyToken(context.Background(), &tokenverify.VerifyTokenRequest{Token: token})
	if err != nil {
		server.logger.Print("failed send auth token : ", err)
		return nil
	}

	server.logger.Print("authorized user : ", authResponse.Username)

	return &userInfo{
		Username: authResponse.Username,
		UserId:   uint(authResponse.UserId),
		Token:    token,
	}

}
