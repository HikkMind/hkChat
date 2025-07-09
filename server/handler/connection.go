package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/structs"
)

func userHandler(connection *websocket.Conn) {
	defer connection.Close()
	defer delete(websocketList, connection)
	for {
		messageType, msg, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		// fmt.Println(string(msg))
		// fmt.Println("got from ", username, " : ", string(message))

		for conn := range websocketList {
			if connection == conn {
				continue
			}
			// data, _ := json.Marshal(structs.Message{Sender: username, Message: string(msg)})
			err = conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func statusAnswer(responseWriter http.ResponseWriter, message string, code int) {
	mess := structs.MessageStatus{Message: message}
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	json.NewEncoder(responseWriter).Encode(mess)
}
