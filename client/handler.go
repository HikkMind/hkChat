package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/hikkmind/hkchat/structs"
)

var (
	username string
)

func handleInput(ctx context.Context, cancelCtx context.CancelFunc, messageChannel chan string) {
	var userInput string
	inputScanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			inputScanner.Scan()
			userInput = inputScanner.Text()
			if len(userInput) == 0 {
				continue
			}

			if userInput == "/exit" {
				cancelCtx()
				return
			} else if userInput == "/login" {
				handleLoginRegister("login")
				break
			} else if userInput == "/register" {
				handleLoginRegister("register")
				break
			} else if userInput == "/messager" {
				go handleConnection(ctx, cancelCtx, messageChannel)
				break
			}

			messageChannel <- userInput

			//connection.Write([]byte(userInput))
		}
	}
}

func handleLoginRegister(operation string) {

	if operation != "register" && operation != "login" {
		fmt.Println("wrong operation : ", operation)
		return
	}

	inputScanner := bufio.NewScanner(os.Stdin)
	fmt.Print("username : ")
	inputScanner.Scan()
	username = inputScanner.Text()

	fmt.Print("password : ")
	inputScanner.Scan()
	password := inputScanner.Text()

	data := structs.AuthUser{Username: username, Password: password}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("GET", "http://localhost:8080/"+operation, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var mess structs.MessageStatus
		json.NewDecoder(resp.Body).Decode(&mess)
		fmt.Println(mess.Message)
		username = ""
	}

}
func handleConnection(ctx context.Context, cancelCtx context.CancelFunc, messageChannel <-chan string) {
	if username == "" {
		fmt.Println("no login")
		return
	}
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/messager", nil)
	defer cancelCtx()
	if err != nil {
		fmt.Println("Failed join to server : ", err)
		return
	}
	defer conn.Close()

	go handleConnectionReceiver(ctx, conn, cancelCtx)
	go handleConnectionSender(ctx, conn, messageChannel)

	<-ctx.Done()

}

func handleConnectionReceiver(ctx context.Context, connection *websocket.Conn, cancelCtx context.CancelFunc) {

	defer cancelCtx()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, msg, _ := connection.ReadMessage()
			if messageType == websocket.CloseMessage {
				return
			}
			var message structs.Message
			err := json.Unmarshal(msg, &message)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(message.Sender, " : ", message.Message)
		}
	}

	// message := make([]byte, 1024)
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	default:
	// 		_, err := connection.Read(message)
	// 		if err != net.ErrClosed && err != nil {
	// 			fmt.Println("Read error : ", err)
	// 			break
	// 		}
	// 		fmt.Println(" server : ", string(message))
	// 	}
	// }

}

func handleConnectionSender(ctx context.Context, connection *websocket.Conn, messageChannel <-chan string) {

	fmt.Println("start sending")
	var userInput string
	for {
		select {
		case <-ctx.Done():
			return
		case userInput = <-messageChannel:
			var message structs.Message
			message.Sender = username
			message.Message = userInput
			data, err := json.Marshal(message)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = connection.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
			// fmt.Println("log : sent ", string(data))
		}
	}

	// var userInput string
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	case userInput = <-messageChannel:
	// 		data := structs.Message{Message: userInput}
	// 		jsonData, _ := json.Marshal(data)
	// 		req, _ := http.NewRequest("POST", "http://localhost:8080/messager", bytes.NewBuffer(jsonData))
	// 		req.Header.Set("Content-Type", "application/json")

	// 		client := &http.Client{}
	// 		client.Do(req)
	// 		fmt.Println("send message : ", userInput)

	// 		//connection.Write([]byte(userInput))
	// 	}
	// }

	// connection.Write([]byte("Hello server!"))

}

func handleSetUsername(connection net.Conn) {
	fmt.Print("Input your username (no spaces) : ")
	fmt.Scan(&username)
	data := structs.AuthUser{Username: username}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("GET", "http://localhost:8080/auth", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error send : ", err)
		return
	}
	defer resp.Body.Close()

}
