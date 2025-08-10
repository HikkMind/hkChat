package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/connection/chat"
	"github.com/hikkmind/hkchat/structs"
)

// type ConnectionChatMessage struct {
// 	Intent string `json:"intent"`
// 	ChatId int    `json:"id"`
// 	// Name         string            `json:"name"`
// 	// UserId       int               `json:"userId"`
// 	// Username     string            `json:"username"`
// 	Message string `json:"message"`
// 	// ChatList     []ChatListMessage `json:"all_chats"`
// 	// UserChannels chat.UserChannels `json:"-"`
// 	Token string `json:"token"`
// }

type ChatInfo struct {
	ChatId   uint   `json:"chat_id"`
	ChatName string `json:"chat_name"`
}

type ChatListInfo struct {
	Intent   string     `json:"intent"`
	Status   string     `json:"status"`
	ChatList []ChatInfo `json:"chat_list"`
}

type ChatMessage struct {
	Intent  string    `json:"intent"`
	Sender  string    `json:"username"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

const (
	JoinChat       = "join_chat"
	LeaveChat      = "leave_chat"
	SendMessage    = "send_message"
	GetChats       = "get_chats"
	GetChatHistory = "get_history"
)

func (server *ChatServer) handleUserConnection(connection *websocket.Conn, currentUser *userInfo) {

	defer connection.Close()

	var outputChannel chan structs.Message

	var message HandleConnectionMessage
	for {
		messageType, msg, err := connection.ReadMessage()
		var sendingContextCancel context.CancelFunc
		var sendingContext context.Context

		if messageType == websocket.CloseMessage || err != nil {
			if sendingContextCancel != nil {
				sendingContextCancel()
			}
			return
		}

		json.Unmarshal(msg, &message)
		server.logger.Print("new chat signal : ", message)
		if message.Intent == JoinChat {

			// inputChannel = make(chan structs.Message)
			outputChannel = make(chan structs.Message)

			sendingContext, sendingContextCancel = context.WithCancel(context.Background())
			go server.handleMessageSending(sendingContext, connection, outputChannel)

			chatChannel, ok := server.chatList[uint(message.ChatId)]
			if ok {
				chatChannel <- chat.ControlMessage{
					Signal: chat.Join,
					UserID: int(currentUser.UserId),
					// InputChannel:  inputChannel,
					OutputChannel: outputChannel,
				}
			} else {
				server.logger.Print("wrong join chat_id")
			}

		} else if message.Intent == LeaveChat {
			server.chatList[uint(message.ChatId)] <- chat.ControlMessage{
				Signal: chat.Leave,
				UserID: int(currentUser.UserId),
			}
			if sendingContextCancel != nil {
				sendingContextCancel()
			}
			// inputChannel = nil
			outputChannel = nil

		} else if message.Intent == SendMessage {
			if len(message.Text) == 0 {
				continue
			}
			chatChannel, ok := server.chatList[uint(message.ChatId)]
			if ok {
				chatChannel <- chat.ControlMessage{
					Signal:   chat.SendMessage,
					UserID:   int(currentUser.UserId),
					Username: currentUser.Username,
					Message:  message.Text,
				}
			} else {
				server.logger.Print("wrong join chat_id")
			}
			// if inputChannel != nil {
			// 	inputChannel <- server.createNewMessage(message, *currentUser)
			// }
		} else if message.Intent == GetChats {
			allChats := make([]ChatInfo, len(server.chatListName))
			ind := 0
			for chatId, chatName := range server.chatListName {
				allChats[ind] = ChatInfo{
					ChatId:   chatId,
					ChatName: chatName,
				}
				ind++
			}

			answer := ChatListInfo{
				Intent:   "chat_list",
				Status:   "ok",
				ChatList: allChats,
			}
			connection.WriteJSON(answer)
		}
	}
}

// func (server *ChatServer) createNewMessage(message HandleConnectionMessage, currentUser userInfo) structs.Message {

// 	messageTime := time.Now()

// 	// var user tables.User
// 	// server.database.First(&user, "id = ?", message.UserId)
// 	tableMessage := tables.Message{
// 		SenderID:  uint(currentUser.UserId),
// 		Message:   message.Text,
// 		CreatedAt: messageTime,
// 	}
// 	server.database.Create(&tableMessage)

// 	newMessage := structs.Message{
// 		Sender:  currentUser.Username,
// 		Message: message.Text,
// 		Time:    messageTime,
// 	}

// 	return newMessage
// }

func (server *ChatServer) handleMessageSending(ctx context.Context, connection *websocket.Conn, outputChannel <-chan structs.Message) {

	// var messageConnection structs.Message

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-outputChannel:
			err := connection.WriteJSON(ChatMessage{
				Intent:  "send_message",
				Sender:  message.Sender,
				Message: message.Message,
				Time:    message.Time,
			})
			if err != nil {
				server.logger.Print("error send message to user : ", err)
			}
		}
	}
}
