package main

import (
	"fmt"

	"github.com/hikkmind/hkchat/server/handler"
)

func main() {
	// var mainServer handler.Server = handler.Server{}
	// handler.MainServer = handler.Server{}
	err := handler.StartServer()
	fmt.Println(err)
	if err != nil {
		fmt.Println("failed to start server : ", err)
		return
	}
	fmt.Println("server started")
}
