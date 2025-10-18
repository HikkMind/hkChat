package server

import (
	"context"
	"errors"
	authstream "hkchat/proto/datastream/auth"
	"hkchat/tables"

	"gorm.io/gorm"
)

func (server *DatabaseServer) VerifyUserPassword(ctx context.Context, request *authstream.UserDataRequest) (*authstream.UserDataResponse, error) {

	server.logger.Print("verify user password...")

	var user tables.User
	result := server.databaseConnection.Where("username = ? AND password = ?", request.Username, request.Password).First(&user)
	verifyResult := &authstream.UserDataResponse{Status: false, UserID: uint32(user.ID)}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		server.logger.Print("wrong login or password")
		return verifyResult, nil
	}
	if result.Error != nil {
		server.logger.Print("request error : ", result.Error.Error())
		return verifyResult, result.Error
	}

	verifyResult.Status = true

	server.logger.Print("user verified")

	return verifyResult, nil
}

func (server *DatabaseServer) RegisterNewUser(ctx context.Context, request *authstream.UserDataRequest) (*authstream.UserDataResponse, error) {
	server.logger.Print("create new user...")
	result := server.databaseConnection.Create(&tables.User{Username: request.Username, Password: request.Password})
	server.logger.Print("new user created")
	return nil, result.Error
}

func (server *DatabaseServer) SetRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {
	server.logger.Print("adding refresh token...")
	err := server.redisConnection.Set(server.redisContext, request.RefreshToken, "", server.refreshTTL).Err()
	server.logger.Print("refresh token added")
	return nil, err
}

func (server *DatabaseServer) UnsetRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {

	server.logger.Print("deleting refresh token...")
	err := server.redisConnection.Del(server.redisContext, request.RefreshToken).Err()
	server.logger.Print("refresh token deleted")
	return nil, err
}

func (server *DatabaseServer) FindRefreshToken(ctx context.Context, request *authstream.UserRefreshTokenRequest) (*authstream.UserRefreshTokenResponse, error) {
	server.logger.Print("search refresh token...")

	// if server.redisConnection == nil {
	// 	server.logger.Fatal("redis connection is nil pointer")
	// }
	exists, err := server.redisConnection.Exists(server.redisContext, request.RefreshToken).Result()

	if err != nil || exists == 0 {
		server.logger.Print("no valid refresh token")
		return &authstream.UserRefreshTokenResponse{
			Status: false,
		}, err
	}

	server.logger.Print("refresh token validated")
	return &authstream.UserRefreshTokenResponse{
		Status: true,
	}, nil
}
