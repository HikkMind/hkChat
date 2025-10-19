package server

import (
	authstream "hkchat/proto/datastream/auth"
	chatstream "hkchat/proto/datastream/chat"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func (server *DatabaseServer) startGrpcServer() {
	grpcServer := grpc.NewServer()
	authstream.RegisterUserDataServiceServer(grpcServer, server)
	chatstream.RegisterChatServiceServer(grpcServer, server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	GRPC_PORT, ok := os.LookupEnv("DATAGATE_GRPC_PORT")
	if !ok {
		server.logger.Fatal("no grpc port in environment")
	}
	if len(GRPC_PORT) == 0 {
		server.logger.Fatal("empty grpc port in environment")
	}

	listener, err := net.Listen("tcp", ":"+GRPC_PORT)
	if err != nil {
		server.logger.Fatal("failed open grpc network : ", err)
		return
	}

	server.logger.Print("start grpc server")

	if err := grpcServer.Serve(listener); err != nil {
		server.logger.Fatal("failed start listen port : ", err)
	}

	server.logger.Print("close grpc server connection")
}
