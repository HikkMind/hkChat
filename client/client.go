package main

import (
	"context"
	"fmt"
	"net"
	// "sync"
)

func main() {

	serverConnection, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("failed to connect server")
		return
	}
	defer serverConnection.Close()
	// handleSetUsername(serverConnection)
	// closeChannel := make(chan struct{})
	globalContext, cancelGlobalContext := context.WithCancel(context.Background())
	messageChannel := make(chan string)
	defer cancelGlobalContext()
	go handleInput(globalContext, cancelGlobalContext, messageChannel)
	go handleConnectionReceiver(globalContext, serverConnection)
	go handleConnectionSender(globalContext, messageChannel)

	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	cancelGlobalContext()
	// }()

	<-globalContext.Done()

}
