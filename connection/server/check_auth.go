package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type authMessage struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func (server *ChatServer) checkAuthToken(token string) *userInfo {

	// checkMessage := authMessage{
	// 	Status: "check_auth",
	// }
	// requestBody, _ := json.Marshal(checkMessage)
	// requestBodyReader := bytes.NewBuffer(requestBody)

	authRequest, err := http.NewRequest("GET", "http://auth:8081/checktoken", nil)
	if err != nil {
		server.logger.Print("Create auth request error:", err)
		return nil
	}
	authRequest.Header.Set("Content-Type", "application/json")
	authRequest.Header.Set("Authorization", "Bearer "+token)

	authClient := http.Client{}
	authResponse, err := authClient.Do(authRequest)
	if err != nil {
		server.logger.Print("Check token error:", err)
		return nil
	}

	if authResponse.StatusCode != http.StatusOK {
		return nil
	}

	var user userInfo
	err = json.NewDecoder(authResponse.Body).Decode(&user)
	if err != nil {
		log.Println("Failed read user info:", err)
		return nil
	}

	return &user
}
