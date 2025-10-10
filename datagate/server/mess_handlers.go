package server

import (
	"context"
	chatstream "hkchat/proto/datastream/chat"
	"hkchat/tables"
)

func (server *DatabaseServer) LoadChatHistory(ctx context.Context, request *chatstream.ChatHistoryRequest) (*chatstream.ChatHistoryResponse, error) {

	chatHistory := &chatstream.ChatHistoryResponse{
		History: []*chatstream.Message{},
	}

	result := server.databaseConnection.
		Table("messages").
		Select("messages.*, users.username").
		Joins("JOIN users ON users.id = messages.sender_id").
		Where("messages.chat_id = ?", request.ChatId).
		Order("messages.created_at ASC").
		Find(&chatHistory.History)
	return chatHistory, result.Error
}

func (server *DatabaseServer) ProcessMessage(ctx context.Context, request *chatstream.MessageTable) (*chatstream.OperationStatus, error) {

	// tables.Message

	result := server.databaseConnection.
		Table("messages").
		Create(tables.Message{
			SenderID:       uint(request.SenderID),
			SenderUsername: request.SenderUsername,
			ChatID:         uint(request.ChatID),
			Message:        request.Message,
		})

	opStatus := &chatstream.OperationStatus{Status: result.Error == nil}

	return opStatus, result.Error
}
