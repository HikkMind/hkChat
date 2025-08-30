package main

import (
	"connection-service/server"
)

func main() {

	var server server.ChatServer

	server.StartServer()
}
