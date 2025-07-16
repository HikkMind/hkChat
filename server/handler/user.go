package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/server/tables"
	"github.com/hikkmind/hkchat/structs"
)

func userHandler(connection *websocket.Conn) {
	defer connection.Close()
	defer delete(mainServer.WebsocketList, connection)

	databaseChannel := make(chan []byte, mainServer.StartMessagesCount)
	messageChannel := mainServer.WebsocketList[connection]

	userContext, userContextCancel := context.WithCancel(context.Background())
	databaseContext, databaseContextCancel := context.WithCancel(userContext)
	defer userContextCancel()

	go getLastMessages(userContext, mainServer.StartMessagesCount, databaseChannel)
	go getUserMessages(userContext, connection, messageChannel, databaseChannel, databaseContextCancel)

	<-databaseContext.Done()

	for {
		messageType, msg, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}

		var message structs.Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println("message wrong json")
			return
		}

		var user tables.User
		mainServer.Database.First(&user, "username = ?", message.Sender)
		mainServer.Database.Create(&tables.Message{SenderID: user.ID, Message: message.Message})

		for conn, channel := range mainServer.WebsocketList {
			if connection == conn {
				continue
			}
			channel <- msg
		}
	}

}

// func statusAnswer(responseWriter http.ResponseWriter, message string, code int) {
// 	mess := structs.MessageStatus{Message: message}
// 	responseWriter.Header().Set("Content-Type", "application/json")
// 	responseWriter.WriteHeader(code)
// 	json.NewEncoder(responseWriter).Encode(mess)
// }

func getLastMessages(userContext context.Context, count int, databaseChannel chan<- []byte) {

	defer close(databaseChannel)
	var messages []structs.Message

	mainServer.Database.Table("messages").
		Select("users.username AS sender, messages.message, messages.created_at AS time").
		Joins("left join users on users.id = messages.sender_id").
		Order("messages.created_at ASC").
		Limit(count).
		Scan(&messages)

	for i := range messages {
		select {
		case <-userContext.Done():
			return
		default:
			msg, _ := json.Marshal(messages[i])
			databaseChannel <- msg
		}
	}
}

func getUserMessages(userContext context.Context, connection *websocket.Conn,
	messageChannel <-chan []byte, databaseChannel <-chan []byte, databaseContextCancel context.CancelFunc) {

	var msg []byte
	var ok bool

	var message structs.Message
	var lastTime time.Time

database_loop:
	for {
		select {
		case <-userContext.Done():
			return
		default:
			msg, ok = <-databaseChannel
			if !ok {
				break database_loop
			}
			err := connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println(err)
			}
			err = json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println(err)
			}
			lastTime = message.Time
		}
	}

	databaseContextCancel()

	for {
		select {
		case <-userContext.Done():
			return
		default:
			msg, ok = <-messageChannel
			if !ok {
				return
			}
			err := json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println(err)
			}

			if message.Time.Before(lastTime) {
				break
			}

			err = connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}
