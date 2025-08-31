package server

import (
	authstream "hkchat/proto/datastream/auth"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func (server *DatabaseServer) startGrpcServer() {
	grpcServer := grpc.NewServer()
	authstream.RegisterUserDataServiceServer(grpcServer, server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", ":6002")
	if err != nil {
		server.logger.Fatal("failed open grpc network : ", err)
		return
	}

	server.logger.Print("create grpc connection")

	if err := grpcServer.Serve(listener); err != nil {
		server.logger.Fatal("failed start listen port : ", err)
	}

	server.logger.Print("close grpc server connection")
}
