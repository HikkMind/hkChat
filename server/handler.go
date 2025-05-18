package main

import (
	"fmt"
	"net"
)

// func handleConnection(connection net.Conn) {

// }

func handleConnectionReader(connection net.Conn, username string) {

	clientMessage := make([]byte, 1024)
	connection.Write([]byte("You was connected with username : " + username))
	defer connection.Close()

	for {
		n, err := connection.Read(clientMessage)
		if err != nil {
			fmt.Printf("client %s disconnected\n", username)
			break
		}
		fmt.Printf("client %s sent message: %s\n", username, string(clientMessage[:n]))
	}
}

// func handleSendMessage(connection net.Conn, message []byte) {
// 	connection.Write(message)
// }
