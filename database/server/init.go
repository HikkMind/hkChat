package server

import (
	authstream "hkchat/proto/datastream/auth"
	"log"
	"os"

	"hkchat/tables"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseServer struct {
	logger         *log.Logger
	databaseClient authstream.UserDataServiceClient

	databaseConnection *gorm.DB
	authstream.UnimplementedUserDataServiceServer
}

func NewServer() *DatabaseServer {
	server := &DatabaseServer{}

	server.logger = log.Default()
	server.logger.SetPrefix("[ DATABASE ]")

	dsn := os.Getenv("DB_CONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	if err := db.AutoMigrate(&tables.User{}, &tables.Chat{}, &tables.Message{}); err != nil {
		log.Fatal("migration failed:", err)
	}
	server.databaseConnection = db
	server.logger.Print("connected to database")

	return server
}

func (server *DatabaseServer) StartServer() {
	go server.startGrpcServer()
}
