package tables

type User struct {
	Id       int    `gorm:"PrimaryKey;autoIncrement"`
	Username string `gorm:"unique;size:20;not null"`
	Password string `gorm:"not null"`
}
