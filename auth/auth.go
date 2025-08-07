package main

import "github.com/hikkmind/hkchat/auth/server"

func main() {

	authServer := server.AuthServer{}
	authServer.StartServer()

}
