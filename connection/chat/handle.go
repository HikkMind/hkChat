package chat

import (
	"context"
	"time"

	"hkchat/tables"

	"gorm.io/gorm"
)

func HandleChat(signalChannel <-chan ControlMessage, chatId uint, database *gorm.DB) {

	currentChat := newChat(chatId, database)
	if currentChat == nil {
		return
	}

	chatContext, chatContextCancel := context.WithCancel(context.Background())
	defer chatContextCancel()

	go currentChat.handleInputMessages(chatContext)

	for {
		signal := <-signalChannel

		currentChat.logger.Print("new signal in chat : ", chatId, ": ", signal)

		if signal.Signal == Join {
			currentChat.userMutex.Lock()
			currentChat.userChannelList[signal.UserID] = signal.OutputChannel
			currentChat.messageMutex.RLock()
			for _, message := range currentChat.messages {
				signal.OutputChannel <- message
			}
			currentChat.messageMutex.RUnlock()
			currentChat.userMutex.Unlock()
			currentChat.logger.Print("user joined : ", signal.Username)

		} else if signal.Signal == Leave {
			currentChat.userMutex.Lock()
			delete(currentChat.userChannelList, signal.UserID)
			currentChat.userMutex.Unlock()
			currentChat.logger.Print("user leaved : ", signal.Username)

		} else if signal.Signal == SendMessage {
			currentChat.logger.Print("start process message in channel")
			currentChat.messageChannel <- tables.Message{
				SenderID:       uint(signal.UserID),
				SenderUsername: signal.Username,
				ChatID:         chatId,
				Message:        signal.Message,
				CreatedAt:      time.Now(),
			}
		}
	}
}
