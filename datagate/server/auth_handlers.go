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

func (server *DatabaseServer) SetRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {
	err := server.redisConnection.Set(server.redisContext, request.RefreshToken, "", server.refreshTTL).Err()
	return nil, err
}

func (server *DatabaseServer) UnsetRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {
	err := server.redisConnection.Del(server.redisContext, request.RefreshToken).Err()
	return nil, err
}

func (server *DatabaseServer) FindRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {
	exists, err := server.redisConnection.Exists(server.redisContext, request.RefreshToken).Result()

	if err != nil || exists == 0 {
		return &authstream.UserRefreshTokenResponse{
			Status: false,
		}, err
	}

	return &authstream.UserRefreshTokenResponse{
		Status: true,
	}, nil
}
