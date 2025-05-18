package main

import (
	"context"
	"fmt"
	"net"
	"time"
	// "sync"
)

func main() {

	// message_to_server := "Hello, Server!"

	serverConnection, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("failed to connect server")
		return
	}
	defer serverConnection.Close()
	handleSetUsername(serverConnection)
	// closeChannel := make(chan struct{})
	globalContext, cancelGlobalContext := context.WithCancel(context.Background())
	defer cancelGlobalContext()
	go handleConnectionReceiver(globalContext, serverConnection)
	go handleConnectionSender(globalContext, serverConnection)

	go func() {
		time.Sleep(10 * time.Second)
		cancelGlobalContext()
	}()

	<-globalContext.Done()

}
