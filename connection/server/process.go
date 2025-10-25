package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"connection-service/chat"
	chatstream "hkchat/proto/datastream/chat"
	"hkchat/structs"

	"github.com/gorilla/websocket"
)

type ChatInfo struct {
	ChatId    uint   `json:"chat_id"`
	ChatName  string `json:"chat_name"`
	OwnerId   uint   `json:"owner_id"`
	OwnerName string `json:"owner_name"`
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
	CreateChat     = "create_chat"
	DeleteChat     = "delete_chat"
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

		switch message.Intent {
		case JoinChat:
			sendingContextCancel = server.handleUserJoin(&connection, uint(message.ChatId), currentUser)
			server.logger.Print("joined user {", currentUser.Username, "}")
		case LeaveChat:
			server.handleUserLeave(sendingContextCancel, uint(message.ChatId), int(currentUser.UserId))
			server.logger.Print("leaved user {", currentUser.Username, "}")
		case SendMessage:
			server.handleUserSendMessage(message, currentUser)
		case GetChats:
			server.handleUserGetChats(connection)
		case CreateChat:
			server.handleCreateChat(currentUser, message.Text)
		case DeleteChat:
			server.handleDeleteChat(currentUser, message.Text)
		default:
			server.logger.Print("unknown intent : ", message)
		}
	}
}

func (server *ChatServer) handleUserJoin(websocketConnection **websocket.Conn, chatId uint, currentUser *userInfo) context.CancelFunc {
	chatContext, chatContextCancel := context.WithCancel(context.Background())
	outputChannel := make(chan structs.Message)
	currentChat, ok := server.chatList[chatId]
	go server.handleConnectionMessageSending(chatContext, websocketConnection, outputChannel)
	if ok {
		currentChat.ControlChannel <- chat.ControlMessage{
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
	server.chatList[chatId].ControlChannel <- chat.ControlMessage{
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
	currentChat, ok := server.chatList[uint(userMessage.ChatId)]
	if ok {
		currentChat.ControlChannel <- chat.ControlMessage{
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
	allChats := make([]ChatInfo, len(server.chatList))
	ind := 0
	for chatId, currentChat := range server.chatList {
		allChats[ind] = ChatInfo{
			ChatId:    chatId,
			ChatName:  currentChat.ChatName,
			OwnerId:   currentChat.OwnerID,
			OwnerName: currentChat.OwnerName,
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

func (server *ChatServer) handleCreateChat(currentUser *userInfo, chatName string) {

	result, err := server.messageDatabaseClient.CreateNewChat(context.Background(), &chatstream.CreateChatRequest{
		UserId:   uint32(currentUser.UserId),
		ChatName: chatName,
	})

	if err != nil || !result.Status {
		server.logger.Printf("failed create chat %s by user %s(%d)\n : %s",
			chatName, currentUser.Username, currentUser.UserId, err)
	}

	server.registerNewChat(uint(result.ChatId), currentUser.UserId, chatName, currentUser.Username)

	// server.logger.Printf("chat %s created by user %s(%d)\n", chatName, currentUser.Username, currentUser.UserId)

}

func (server *ChatServer) handleDeleteChat(currentUser *userInfo, stringChatId string) {

	chatId, err := strconv.Atoi(stringChatId)
	if err != nil {
		server.logger.Print("failed convert chat id for delete : ", stringChatId)
		return
	}

	if server.chatList[uint(chatId)].OwnerID != currentUser.UserId {
		server.logger.Printf("wrong chat owner : %s doesnt owns %s\n", currentUser.Username, stringChatId)
		return
	}

	opStatus, _ := server.messageDatabaseClient.DeleteChat(context.Background(), &chatstream.ChatIdRequest{ChatId: int32(chatId)})
	if opStatus.Status {
		server.chatListMutex.Lock()
		delete(server.chatList, uint(chatId))
		server.chatListMutex.Unlock()
	}
	// server.logger.Printf("user %s deleting chat %s\n", currentUser.Username, stringChatId)
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
