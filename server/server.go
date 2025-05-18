package main

import (
	"fmt"
	"net"
)

func main() {

	listenConnection, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("failed to start listening")
		return
	}
	defer listenConnection.Close()
	userConnection := make(map[string]net.Conn)

	// usernames := make(map[net.Conn]string)

	fmt.Printf("start listening by %s\n", listenConnection)

	messageBuffer := make([]byte, 256)
	for {
		var username string
		connection, err := listenConnection.Accept()
		if err != nil {
			fmt.Println("failed get connection")
			break
		}

		n, _ := connection.Read(messageBuffer)
		username = string(messageBuffer[:n])
		userConnection[username] = connection

		// handleSendMessage(connection, []byte("hi from server"))
		go handleConnectionReader(connection, username)
	}

}
