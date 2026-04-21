package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"connection-service/chat"

	chatstream "hkchat/proto/datastream/chat"
	"hkchat/structs"
	"hkchat/tables"

	tokenverify "github.com/hikkmind/hkchat/proto/tokenverify"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type chatControlInfo struct {
	ControlChannel chan structs.ControlMessage
	ChatName       string
	OwnerID        uint
	OwnerName      string
}

type ChatListSignal struct {
	Intent         string   `json:"intent"`
	ChatParameters ChatInfo `json:"chat_info"`
	// ChatId uint   `json:"chat_id"`
}

type ChatServer struct {
	chatList map[uint]chatControlInfo
	// chatList     map[uint]chan chat.ControlMessage
	// chatListName map[uint]string
	// chatListChannel chan chatstream.ChatInfo
	chatListMutex sync.RWMutex

	userChatSignal          map[uint]chan ChatListSignal
	serverChatSignalChannel chan ChatListSignal
	userChatSignalMutex     sync.RWMutex

	logger *log.Logger

	serverPort   string
	datagatePort string
	authPort     string

	authTokenClient         tokenverify.AuthServiceClient
	messageDatabaseClient   chatstream.ChatServiceClient
	chatGlobalContext       context.Context
	chatGlobalContextCancel context.CancelFunc

	rabbitConn     *amqp.Connection
	rabbitChannel  *amqp.Channel
	rabbitExchange string
}

type HandleConnectionMessage struct {
	Intent string `json:"intent"`
	Status string `json:"status"`
	ChatId int    `json:"chat_id"`
	Token  string `json:"token"`
	Text   string `json:"text"`
}

type userInfo struct {
	Username string `json:"username"`
	UserId   uint   `json:"user_id"`
	Token    string `json:"token"`
}

func (server *ChatServer) StartServer() {
	server.serverVariablesInit()
	serverHTTP := &http.Server{
		Addr: server.serverPort,
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	http.HandleFunc("/chatlist", server.connectUser)

	server.grpcDatagateInit()
	server.grpcAuthInit()
	server.InitRabbitMQ()

	server.loadChatList()
	go server.startHandleChatSignal()

	serverHTTP.ListenAndServe()

	server.logger.Print("server stopped")

}

func (server *ChatServer) grpcAuthInit() {
	tokenConnection, err := grpc.NewClient("auth"+server.authPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.logger.Print("failed connect auth grpc : ", err)
		return
	}
	server.authTokenClient = tokenverify.NewAuthServiceClient(tokenConnection)
	server.logger.Print("connected to auth grpc server")
}

func (server *ChatServer) grpcDatagateInit() {
	dataConnection, err := grpc.NewClient("datagate"+server.datagatePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		server.logger.Print("failed connect datagate grpc : ", err)
		return
	}
	server.messageDatabaseClient = chatstream.NewChatServiceClient(dataConnection)
	server.logger.Print("connected to datagate grpc server")
}

func (server *ChatServer) loadChatList() {

	server.logger.Print("loading chat list...")

	allChats, err := server.messageDatabaseClient.LoadChatList(context.Background(), &chatstream.ChatListRequest{})

	if err != nil {
		server.logger.Fatal("failed load chats : ", err)
		return
	}

	server.logger.Print("loaded chats : ", len(allChats.ChatList))
	server.logger.Print("chat list: ", allChats)

	server.chatList = make(map[uint]chatControlInfo)

	for _, currentChat := range allChats.ChatList {
		server.registerNewChat(uint(currentChat.ChatID), uint(currentChat.OwnerID), currentChat.ChatName, currentChat.OwnerName)
	}

	server.logger.Print("start handle all chats")
}

func (server *ChatServer) serverVariablesInit() {

	server.serverPort = ":" + os.Getenv("CONN_PORT")
	server.authPort = ":" + os.Getenv("AUTH_GRPC_PORT")
	server.datagatePort = ":" + os.Getenv("DATAGATE_GRPC_PORT")

	// server.chatListChannel = make(chan chatstream.ChatInfo)
	server.userChatSignal = make(map[uint]chan ChatListSignal)
	server.serverChatSignalChannel = make(chan ChatListSignal)

	server.chatGlobalContext, server.chatGlobalContextCancel = context.WithCancel(context.Background())

	server.logger = log.Default()
	server.logger.SetPrefix("[ CONNECTION ]")
}

func (server *ChatServer) registerNewChat(chatID, ownerId uint, chatName, ownerName string) {

	server.chatListMutex.Lock()
	defer server.chatListMutex.Unlock()

	chatChannel := make(chan structs.ControlMessage)

	server.chatList[chatID] = chatControlInfo{
		ControlChannel: chatChannel,
		ChatName:       chatName,
		OwnerID:        ownerId,
		OwnerName:      ownerName,
	}

	go chat.HandleChat(chatChannel, chatID)
}

func (server *ChatServer) startHandleChatSignal() {
	for {
		select {
		case <-server.chatGlobalContext.Done():
			return
		case chatSignal := <-server.serverChatSignalChannel:
			server.userChatSignalMutex.RLock()
			for _, signalChannel := range server.userChatSignal {
				signalChannel <- chatSignal
			}
			server.userChatSignalMutex.RUnlock()
		}
	}
}

func (server *ChatServer) InitRabbitMQ() {
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
		os.Getenv("RABBITMQ_VHOST"),
	)
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		// return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		server.logger.Fatal("failed init rabbitMQ : ", err)
	}
	server.rabbitConn = conn

	ch, err := conn.Channel()
	if err != nil {
		// return err
		server.logger.Fatal("failed get rabbitMQ channel : ", err)
	}
	server.rabbitChannel = ch

	// Объявляем topic exchange (будет использоваться и бекендом, и датагейтом)
	server.rabbitExchange = "chat.events"
	err = ch.ExchangeDeclare(server.rabbitExchange, "topic", true, false, false, false, nil)
	if err != nil {
		// return fmt.Errorf("exchange declare failed: %w", err)
		server.logger.Fatal("exchange declare failed: ", err)
	}

	// Настройка подтверждений publisher confirms
	if err := ch.Confirm(false); err != nil {
		// return fmt.Errorf("confirm mode failed: %w", err)
		server.logger.Fatal("confirm mode failed: ", err)
	}

	log.Println("RabbitMQ initialized (publisher mode)")
	// return nil
}

// publishEvent отправляет событие в RabbitMQ
func (s *ChatServer) publishMessage(chatID uint, msg tables.Message) error {
	s.logger.Print("start publishing message : ", msg)
	envelope := structs.NewEnvelope(chatID, msg)
	body, err := envelope.ToJSON()
	if err != nil {
		return err
	}

	// routing key: событие направляется в датагейт
	routingKey := fmt.Sprintf("chat.%d.event", chatID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.rabbitChannel.PublishWithContext(
		ctx,
		s.rabbitExchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			MessageId:    envelope.EventID,
			Timestamp:    envelope.CreatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	log.Printf("Event published: chat=%d, type=%s, event_id=%s", chatID, "send message", envelope.EventID)
	return nil
}
