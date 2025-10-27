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
	CreateChat
	DeleteChat
)

type ControlMessage struct {
	Signal        ChatSignal
	UserID        int
	Username      string
	Message       string
	OutputChannel chan structs.Message
}

type Chat struct {
	userChannelList map[int]chan structs.Message
	messages        []structs.Message
	messageChannel  chan tables.Message

	userMutex    sync.RWMutex
	messageMutex sync.RWMutex

	chatId uint
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
	self.logger = log.Default()
	self.logger.SetPrefix("[ CHAT ]")
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

	chatHistory, err := currentChat.databaseClient.LoadChatHistory(context.Background(), &chatstream.ChatIdRequest{
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
