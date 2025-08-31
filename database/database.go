package main

import (
	"database-service/server"
)

func main() {

	databaseServer := server.NewServer()
	databaseServer.StartServer()
}
