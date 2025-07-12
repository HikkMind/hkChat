package tables

type User struct {
	ID       uint   `gorm:"PrimaryKey;autoIncrement"`
	Username string `gorm:"unique;size:20;not null"`
	Password string `gorm:"not null"`
	Messages string `gorm:"foreignKey:SenderID;references:ID;constraint:OnDelete:CASCADE"`
}

type Message struct {
	ID       uint `gorm:"PrimaryKey;autoIncrement"`
	SenderID uint
	Message  string
}
