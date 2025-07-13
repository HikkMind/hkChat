package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/server/tables"
	"github.com/hikkmind/hkchat/structs"
)

func userHandler(connection *websocket.Conn) {
	defer connection.Close()
	defer delete(websocketList, connection)

	databaseChannel := make(chan []byte, startMessagesCount)
	messageChannel := websocketList[connection]

	userContext, userContextCancel := context.WithCancel(context.Background())
	defer userContextCancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go getLastMessages(userContext, startMessagesCount, databaseChannel)
	go getUserMessages(userContext, connection, messageChannel, databaseChannel, &wg)

	wg.Wait()

	for {
		messageType, msg, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		// fmt.Println(string(msg))
		// fmt.Println("got from ", username, " : ", string(message))

		var message structs.Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println("message wrong json")
			return
		}

		var user tables.User
		database.First(&user, "username = ?", message.Sender)
		database.Create(&tables.Message{SenderID: user.ID, Message: message.Message})

		for conn, channel := range websocketList {
			if connection == conn {
				continue
			}
			// data, _ := json.Marshal(structs.Message{Sender: username, Message: string(msg)})
			// err = conn.WriteMessage(websocket.TextMessage, msg)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			channel <- msg
		}
	}

}

func statusAnswer(responseWriter http.ResponseWriter, message string, code int) {
	mess := structs.MessageStatus{Message: message}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	json.NewEncoder(responseWriter).Encode(mess)
}

func getLastMessages(userContext context.Context, count int, databaseChannel chan<- []byte) {

	defer close(databaseChannel)
	var messages []structs.Message
	// database.Order("created_at DESC").Limit(count).Find(messages)
	// sort.Slice(messages, func(i, j int) bool {
	// 	return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	// })

	database.Table("messages").
		Select("users.username AS sender, messages.message, messages.created_at AS time").
		Joins("left join users on users.id = messages.sender_id").
		Order("messages.created_at ASC").
		Limit(count).
		Scan(&messages)

	// msg := make([][]byte, count)

	for i := range messages {
		// msg[i], _ = json.Marshal(messages[i])
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
	messageChannel <-chan []byte, databaseChannel <-chan []byte, wg *sync.WaitGroup) {

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

	wg.Done()

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
