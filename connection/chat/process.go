package chat

import (
	"context"

	chatstream "hkchat/proto/datastream/chat"
	"hkchat/structs"
	"hkchat/tables"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (chat *Chat) handleInputMessages(chatContext context.Context) {

	for {
		select {
		case <-chatContext.Done():
			return
		case message := <-chat.messageChannel:
			chat.logger.Print("get new message from message channel : ", message)
			chat.processNewMessage(&message)
		}
	}
}

func (chat *Chat) processNewMessage(message *tables.Message) {

	// result := chat.database.
	// 	Table("messages").
	// 	Create(message)
	_, err := chat.databaseClient.ProcessMessage(context.Background(), &chatstream.MessageTable{
		SenderID:       uint32(message.SenderID),
		SenderUsername: message.SenderUsername,
		ChatID:         uint32(message.ChatID),
		Message:        message.Message,
		Time:           timestamppb.New(message.CreatedAt),
	})

	if err != nil {
		chat.logger.Print("create chat message error for chat ", chat.chatId, ": ", message)
		return
	}
	chat.logger.Print("add new message to database")

	userMessage := structs.Message{
		Sender:  message.SenderUsername,
		Message: message.Message,
		Time:    message.CreatedAt,
	}

	// chat.userMutex.RLock()
	chat.messageMutex.RLock()
	chat.messages = append(chat.messages, userMessage)
	for _, userChannel := range chat.userChannelList {
		userChannel <- userMessage
	}
	// chat.userMutex.RUnlock()
	chat.messageMutex.RUnlock()
	chat.logger.Print("send new message to users")

}
