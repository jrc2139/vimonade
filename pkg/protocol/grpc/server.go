package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	v1 "github.com/jrc2139/vimonade/pkg/api/v1"
)

// RunServer registers gRPC service and run server.
func RunServer(ctx context.Context, srv v1.MessageServiceServer, creds credentials.TransportCredentials, allowRange, serverAddr string) error {
	listen, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}

	// register service
	var server *grpc.Server

	if creds == nil {
		// insecure
		server = grpc.NewServer()
	} else {
		// secure
		server = grpc.NewServer(grpc.Creds(creds))
	}

	v1.RegisterMessageServiceServer(server, srv)

	// start gRPC server
	log.Println("starting gRPC server...")

	return server.Serve(listen)
}
