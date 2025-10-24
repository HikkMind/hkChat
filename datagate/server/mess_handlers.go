package server

import (
	"context"
	chatstream "hkchat/proto/datastream/chat"
	"hkchat/tables"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *DatabaseServer) LoadChatHistory(ctx context.Context, request *chatstream.ChatIdRequest) (*chatstream.ChatHistoryResponse, error) {

	server.logger.Print("loading chat \"", request.ChatId, "\" history...")

	chatHistory := &chatstream.ChatHistoryResponse{
		History: []*chatstream.Message{},
	}

	type messageInfo struct {
		SenderUsername string `gorm:"column:username"`
		Message        string
		CreatedAt      time.Time
	}

	tableChatHistory := make([]messageInfo, 0)

	result := server.databaseConnection.
		Table("messages").
		Select("messages.*, users.username").
		Joins("JOIN users ON users.id = messages.sender_id").
		Where("messages.chat_id = ?", request.ChatId).
		Order("messages.created_at ASC").
		Find(&tableChatHistory)

	chatHistory.History = make([]*chatstream.Message, len(tableChatHistory))
	for ind, message := range tableChatHistory {
		chatHistory.History[ind] = &chatstream.Message{
			Sender:  message.SenderUsername,
			Message: message.Message,
			Time:    timestamppb.New(message.CreatedAt),
		}
	}

	server.logger.Print("loaded chat \"", request.ChatId, "\"")

	return chatHistory, result.Error
}

func (server *DatabaseServer) ProcessMessage(ctx context.Context, request *chatstream.MessageTable) (*chatstream.OperationStatus, error) {

	server.logger.Print("processing new message: ", request, "...")

	newMessage := tables.Message{
		SenderID:       uint(request.SenderID),
		SenderUsername: request.SenderUsername,
		ChatID:         uint(request.ChatID),
		Message:        request.Message,
	}

	result := server.databaseConnection.
		Table("messages").
		Create(&newMessage)

	opStatus := &chatstream.OperationStatus{Status: result.Error == nil}
	server.logger.Print("message chat ID: ", request.ChatID, " processed")

	return opStatus, result.Error
}

func (server *DatabaseServer) LoadChatList(ctx context.Context, request *chatstream.ChatListRequest) (*chatstream.ChatListResponse, error) {

	server.logger.Print("loading chat list...")

	type ChatOwnerTable struct {
		OwnerName string `gorm:"column:owner_name" json:"owner_name"`
		tables.Chat
	}

	allChats := make([]ChatOwnerTable, 0)
	result := server.databaseConnection.
		Table("chats").
		Select("chats.id, chats.name, chats.owner_id, users.username as owner_name").
		Joins("JOIN users ON users.id = chats.owner_id").
		Find(&allChats)

	if result.Error != nil {
		server.logger.Print("failed load chat list : ", result.Error)
		return nil, result.Error
	}

	server.logger.Print("all chats: ", allChats)

	response := make([]*chatstream.ChatInfo, len(allChats))
	for ind, chat := range allChats {
		response[ind] = &chatstream.ChatInfo{
			ChatID:    uint32(chat.ID),
			ChatName:  chat.Name,
			OwnerID:   uint32(chat.OwnerID),
			OwnerName: chat.OwnerName,
		}
	}

	server.logger.Print("chat list loaded")

	return &chatstream.ChatListResponse{
		ChatList: response,
	}, nil
}
func (server *DatabaseServer) CreateNewChat(ctx context.Context, request *chatstream.CreateChatRequest) (*chatstream.CreateChatResponse, error) {

	newChat := tables.Chat{
		Name:    request.ChatName,
		OwnerID: uint(request.UserId),
	}

	result := server.databaseConnection.
		Table("chats").
		Create(&newChat)

	chatCreateInfo := &chatstream.CreateChatResponse{Status: result.Error == nil, ChatId: uint32(newChat.ID)}
	server.logger.Printf("chat %s(%d) created by user (%d)\n", request.ChatName, newChat.ID, request.UserId)

	return chatCreateInfo, result.Error
}

func (server *DatabaseServer) DeleteChat(ctx context.Context, request *chatstream.ChatIdRequest) (*chatstream.OperationStatus, error) {
	server.logger.Printf("chat deleting %d...\n", request.ChatId)

	result := server.databaseConnection.Table("chats").Delete(&tables.Chat{}, request.ChatId)

	if result.Error == nil {
		server.logger.Printf("chat %d deleted.\n", request.ChatId)
	} else {
		server.logger.Printf("failed delete chat %d\n", request.ChatId)
	}

	return &chatstream.OperationStatus{
		Status: result.Error == nil,
	}, result.Error
}
