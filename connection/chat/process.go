package chat

import (
	"context"

	"github.com/hikkmind/hkchat/structs"
	"github.com/hikkmind/hkchat/tables"
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

	result := chat.database.
		Table("messages").
		Create(message)

	if result.Error != nil {
		chat.logger.Print("create chat message error for chat ", chat.chatId, ": ", message)
		return
	}
	chat.logger.Print("add new message to database")

	userMessage := structs.Message{
		Sender:  message.SenderUsername,
		Message: message.Message,
		Time:    message.CreatedAt,
	}

	chat.userMutex.RLock()
	chat.messages = append(chat.messages, userMessage)
	for _, userChannel := range chat.userChannelList {
		userChannel <- userMessage
	}
	chat.userMutex.RUnlock()
	chat.logger.Print("send new message to users")

}
