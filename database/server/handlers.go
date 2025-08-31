package server

import (
	"context"
	"errors"
	authstream "hkchat/proto/datastream/auth"
	"hkchat/tables"

	"gorm.io/gorm"
)

func (server *DatabaseServer) VerifyUserPassword(ctx context.Context, request *authstream.UserDataRequest) (*authstream.UserDataResponse, error) {
	var user tables.User
	result := server.databaseConnection.Where("username = ? AND password = ?", request.Username, request.Password).First(&user)
	verifyResult := &authstream.UserDataResponse{Status: false}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		server.logger.Print("wrong login or password")
		return verifyResult, nil
	}
	if result.Error != nil {
		server.logger.Print("request error : ", result.Error.Error())
		return verifyResult, result.Error
	}

	verifyResult.Status = true

	return verifyResult, nil
}

func (server *DatabaseServer) RegisterNewUser(ctx context.Context, request *authstream.UserDataRequest) (*authstream.UserDataResponse, error) {
	result := server.databaseConnection.Create(&tables.User{Username: request.Username, Password: request.Password})
	return nil, result.Error
}
