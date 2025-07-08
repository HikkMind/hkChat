package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

func main() {

	mainServer := &http.Server{
		Addr: ":8080",
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, "connection", connection)
		},
	}

	handlerInit()

	http.HandleFunc("/messager", messageHandler)
	http.HandleFunc("/login", authLogin)
	http.HandleFunc("/register", authRegister)
	fmt.Println("server started try")
	// err := http.ListenAndServe("localhost:8080", nil)
	err := mainServer.ListenAndServe()
	fmt.Println(err)
	if err != nil {
		fmt.Println("failed to start server : ", err)
		return
	}
	fmt.Println("server started")
}
