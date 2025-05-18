package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
)

func handleConnectionReceiver(ctx context.Context, connection net.Conn) {

	message := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := connection.Read(message)
			if err != nil {
				fmt.Println("Read error : ", err)
				break
			}
			fmt.Println(" server : ", string(message))
		}
	}

}

func handleConnectionSender(ctx context.Context, connection net.Conn) {
	var userInput string
	inputScanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			inputScanner.Scan()
			userInput = inputScanner.Text()
			if userInput == "/exit()" {
				return
			}
			connection.Write([]byte(userInput))
		}
	}
	// connection.Write([]byte("Hello server!"))

}

func handleSetUsername(connection net.Conn) {
	var username string
	fmt.Print("Input your username (no spaces) : ")
	fmt.Scan(&username)
	connection.Write([]byte(username))
}
