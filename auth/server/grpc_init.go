package server

import (
	"net"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	tokenverify "github.com/hikkmind/hkchat/proto/tokenverify"
	"google.golang.org/grpc"
)

func (server *AuthServer) startGrpcServer() {

	grpcServer := grpc.NewServer()
	tokenverify.RegisterAuthServiceServer(grpcServer, server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", ":6001")
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
