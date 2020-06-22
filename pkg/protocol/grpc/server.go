package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/pocke/go-iprange"

	v1 "github.com/jrc2139/vimonade/pkg/api/v1"
)

var ra *iprange.Range

// RunServer registers gRPC service and run server.
func RunServer(ctx context.Context, srv v1.MessageServiceServer, creds credentials.TransportCredentials, allowRange, serverAddr string) error {
	listen, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}

	ra, err = iprange.New(allowRange)
	if err != nil {
		return err
	}

	// register service
	var server *grpc.Server

	if creds == nil {
		// insecure
		server = grpc.NewServer(grpc.UnaryInterceptor(IpInterceptor))
	} else {
		// secure
		server = grpc.NewServer(grpc.Creds(creds))
	}

	v1.RegisterMessageServiceServer(server, srv)

	// start gRPC server
	log.Println("starting gRPC server...")

	return server.Serve(listen)
}

func IpInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p, _ := peer.FromContext(ctx)
	if !ra.IncludeStr(p.Addr.String()) {
		return nil, errors.New(fmt.Sprintf("not in allow ip range: %s", p.Addr.String()))
		// http.Error(w, "Not allow ip.", 503)
		// logger.Info("not in allow ip. from: ", p.Addr.String())
		// return
	}
	// Calls the handler
	return handler(ctx, req)

	// Invoke 'handler' to use your gRPC server implementation and get
	// the response.
}
