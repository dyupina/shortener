package grpc

import (
	"net"
	pb "shortener/internal/domain/models/proto"
	"shortener/internal/handlers"

	"google.golang.org/grpc"
)

// RunGRPCServer starts gRPC service.
func RunGRPCServer(con *handlers.Controller) {
	gRPCPort := "3200"
	grpcServer := grpc.NewServer()
	pb.RegisterURLShortenerServer(grpcServer, NewGRPCServer(con))

	lis, err := net.Listen("tcp", ":"+gRPCPort)
	if err != nil {
		con.Logger.Errorf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		con.Logger.Infof("Starting gRPC server on port %s", gRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			con.Logger.Errorf("Failed to start gRPC server: %v", err)
		}
	}()
}
