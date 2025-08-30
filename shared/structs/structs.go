package structs

import "time"

type Message struct {
	Sender  string    `gorm:"column:username" json:"username"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type AuthUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MessageStatus struct {
	Message string `json:"error_message"`
}
