package chat

import (
	"context"
	"time"

	"hkchat/tables"
)

func HandleChat(signalChannel <-chan ControlMessage, chatId uint) {

	currentChat := newChat(chatId)
	if currentChat == nil {
		return
	}

	chatContext, chatContextCancel := context.WithCancel(context.Background())
	defer chatContextCancel()

	go currentChat.handleInputMessages(chatContext)

	for {
		signal := <-signalChannel

		currentChat.logger.Print("new signal in chat : ", chatId, ": ", signal)

		switch signal.Signal {
		case Join:
			currentChat.userMutex.Lock()
			currentChat.userChannelList[signal.UserID] = signal.OutputChannel
			currentChat.messageMutex.RLock()
			for _, message := range currentChat.messages {
				signal.OutputChannel <- message
			}
			currentChat.messageMutex.RUnlock()
			currentChat.userMutex.Unlock()
			currentChat.logger.Print("user joined : ", signal.Username)
		case Leave:
			currentChat.userMutex.Lock()
			delete(currentChat.userChannelList, signal.UserID)
			currentChat.userMutex.Unlock()
			currentChat.logger.Print("user leaved : ", signal.Username)
		case SendMessage:
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
