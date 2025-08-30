package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"connection-service/chat"
	"hkchat/structs"

	"github.com/gorilla/websocket"
)

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
	defer server.logger.Println("finish handle connection for :", currentUser.UserId, currentUser.Username)

	var sendingContextCancel context.CancelFunc
	var message HandleConnectionMessage

	for {
		messageType, msg, err := connection.ReadMessage()

		if messageType == websocket.CloseMessage || err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				if sendingContextCancel != nil {
					sendingContextCancel()
				}
				server.logger.Print("websocket close for user : ", currentUser.Username)
				return
			}
			server.logger.Print("websocket close by error code : ", websocket.CloseMessage)
			return
		}

		json.Unmarshal(msg, &message)
		server.logger.Print("new chat signal : ", message)

		if message.Intent == JoinChat {
			sendingContextCancel = server.handleUserJoin(&connection, uint(message.ChatId), currentUser)
			server.logger.Print("joined user {", currentUser.Username, "}")

		} else if message.Intent == LeaveChat {
			server.handleUserLeave(sendingContextCancel, uint(message.ChatId), int(currentUser.UserId))
			server.logger.Print("leaved user {", currentUser.Username, "}")

		} else if message.Intent == SendMessage {
			server.handleUserSendMessage(message, currentUser)
		} else if message.Intent == GetChats {
			server.handleUserGetChats(connection)
		}
	}
}

func (server *ChatServer) handleUserJoin(websocketConnection **websocket.Conn, chatId uint, currentUser *userInfo) context.CancelFunc {
	chatContext, chatContextCancel := context.WithCancel(context.Background())
	outputChannel := make(chan structs.Message)
	chatChannel, ok := server.chatList[chatId]
	go server.handleConnectionMessageSending(chatContext, websocketConnection, outputChannel)
	if ok {
		chatChannel <- chat.ControlMessage{
			Signal:        chat.Join,
			UserID:        int(currentUser.UserId),
			OutputChannel: outputChannel,
		}
	} else {
		server.logger.Print("wrong join chat_id")
	}

	return chatContextCancel
}

func (server *ChatServer) handleUserLeave(chatContextCancel context.CancelFunc, chatId uint, userId int) {
	server.chatList[chatId] <- chat.ControlMessage{
		Signal: chat.Leave,
		UserID: userId,
	}
	if chatContextCancel != nil {
		chatContextCancel()
	}
}

func (server *ChatServer) handleUserSendMessage(userMessage HandleConnectionMessage, currentUser *userInfo) bool {
	if len(userMessage.Text) == 0 {
		return false
	}
	chatChannel, ok := server.chatList[uint(userMessage.ChatId)]
	if ok {
		chatChannel <- chat.ControlMessage{
			Signal:   chat.SendMessage,
			UserID:   int(currentUser.UserId),
			Username: currentUser.Username,
			Message:  userMessage.Text,
		}
		return true
	} else {
		server.logger.Print("wrong join chat_id : ", userMessage.ChatId)
	}
	return false
}

func (server *ChatServer) handleUserGetChats(websocketConnection *websocket.Conn) {
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
	websocketConnection.WriteJSON(answer)
}

func (server *ChatServer) handleConnectionMessageSending(ctx context.Context, connection **websocket.Conn, outputChannel <-chan structs.Message) {

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-outputChannel:
			err := (*connection).WriteJSON(ChatMessage{
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

func (server *ChatServer) newWebsocketConnection(responseWriter http.ResponseWriter, request *http.Request) (*userInfo, *websocket.Conn) {
	websocketUpgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	server.logger.Print("try connect websocket")

	websocketConnection, err := websocketUpgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		server.logger.Print("failed upgrade websocket : ", err)
		return nil, nil
	}
	server.logger.Print("connected new websocket")

	return server.verifyUserToken(websocketConnection), websocketConnection
}
