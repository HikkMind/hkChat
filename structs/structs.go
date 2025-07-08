package structs

type Message struct {
	// Sender  string `json:"username"`
	Message string `json:"message"`
}

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MessageStatus struct {
	Message string `json:"error_message"`
}
