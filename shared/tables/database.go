package tables

import "time"

type User struct {
	ID       uint      `gorm:"PrimaryKey;autoIncrement"`
	Username string    `gorm:"unique;size:20;not null"`
	Password string    `gorm:"not null"`
	Messages []Message `gorm:"foreignKey:SenderID;references:ID;constraint:OnDelete:CASCADE"`
	Chats    []Chat    `gorm:"foreignKey:OwnerID;references:ID;constraint:OnDelete:CASCADE"`
}

type Message struct {
	ID             uint `gorm:"PrimaryKey;autoIncrement"`
	SenderID       uint
	SenderUsername string `gorm:"-"`
	ChatID         uint   `gorm:"not null"`
	Message        string
	CreatedAt      time.Time `json:"time"`
}

type Chat struct {
	ID        uint      `gorm:"PrimaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"-"`
	OwnerID   uint      `gorm:"not null"`
	Messages  []Message `gorm:"foreignKey:ChatID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Chat) TableName() string {
	return "chats"
}
