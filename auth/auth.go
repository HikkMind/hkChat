package main

import "auth-service/server"

func main() {

	authServer := server.AuthServer{}
	authServer.StartServer()

}
