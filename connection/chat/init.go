package chat

import (
	"context"
	"log"
	"os"
	"sync"

	"hkchat/structs"
	"hkchat/tables"

	chatstream "hkchat/proto/datastream/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatSignal int

const (
	Join ChatSignal = iota
	Leave
	SendMessage
)

type ControlMessage struct {
	Signal   ChatSignal
	UserID   int
	Username string
	// InputChannel  chan structs.Message
	Message       string
	OutputChannel chan structs.Message
}

// type ChatMessage struct {
// 	UserId   uint      `json:"user_id"`
// 	Username string    `json:"username"`
// 	Message  string    `json:"message"`
// 	Time     time.Time `json:"time"`
// }

// type UserChannels struct {
// 	InputChannel  chan structs.Message
// 	OutputChannel chan structs.Message
// }

type Chat struct {
	// userList map[int]UserChannels
	userChannelList map[int]chan structs.Message
	messages        []structs.Message
	messageChannel  chan tables.Message

	userMutex    sync.RWMutex
	messageMutex sync.RWMutex

	chatId uint
	// database *gorm.DB
	logger *log.Logger

	datagatePort   string
	databaseClient chatstream.ChatServiceClient
}

func newChat(chatId uint) *Chat {

	var self Chat
	self.userChannelList = make(map[int]chan structs.Message)
	self.messageChannel = make(chan tables.Message)
	self.messages = make([]structs.Message, 0)
	self.chatId = chatId
	// self.database = database
	self.logger = log.Default()
	self.logger.SetPrefix("[ CHAT ]")
	// self.databaseInit()
	self.datagateGrpcConnectionInit()

	err := self.loadChatHistory()
	if err != nil {
		self.logger.Print("failed load chat history for ", chatId, ": ", err.Error())
		return nil
	}

	self.logger.Print("handle chat : ", chatId)
	return &self
}

func (currentChat *Chat) loadChatHistory() error {
	// result := currentChat.database.
	// 	Table("messages").
	// 	Select("messages.*, users.username").
	// 	Joins("JOIN users ON users.id = messages.sender_id").
	// 	Where("messages.chat_id = ?", currentChat.chatId).
	// 	Order("messages.created_at ASC").
	// 	Find(&currentChat.messages)

	chatHistory, err := currentChat.databaseClient.LoadChatHistory(context.Background(), &chatstream.ChatHistoryRequest{
		ChatId: int32(currentChat.chatId),
	})

	if err != nil {
		return err
	}

	currentChat.messages = make([]structs.Message, len(chatHistory.History))
	for i := range len(chatHistory.History) {
		currentChat.messages[i] = structs.Message{
			Sender:  chatHistory.History[i].Sender,
			Message: chatHistory.History[i].Message,
			Time:    chatHistory.History[i].Time.AsTime(),
		}
	}

	currentChat.logger.Print("got chat history of len : ", len(currentChat.messages))
	if len(currentChat.messages) > 0 {
		currentChat.logger.Print("first history message : ", currentChat.messages[0])
	}

	return err
}

// func (currentChat *Chat) databaseInit() {
// 	_, err := grpc.NewClient("datagate"+datagatePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		server.logger.Print("failed check auth token : ", err)
// 		return
// 	}
// 	server.databaseClient = authstream.NewUserDataServiceClient(tokenConnection)
// 	server.logger.Print("connected to grpc server")
// }

func (currentChat *Chat) datagateGrpcConnectionInit() {
	datagatePort := ":" + os.Getenv("DATAGATE_GRPC_PORT")

	dataConnection, err := grpc.NewClient("datagate"+datagatePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		currentChat.logger.Print("failed connect to datagate : ", err)
		return
	}
	currentChat.databaseClient = chatstream.NewChatServiceClient(dataConnection)
	currentChat.logger.Print("connected to grpc server")
}
