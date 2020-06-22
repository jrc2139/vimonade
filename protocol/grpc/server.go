package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	log "github.com/inconshreveable/log15"
	"github.com/pocke/go-iprange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	pb "github.com/jrc2139/vimonade/api"
)

// RunServer registers gRPC service and run server.
func RunServer(ctx context.Context, srv pb.VimonadeServiceServer, logger log.Logger, creds credentials.TransportCredentials, allowRange, serverAddr string) error {
	listen, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}

	ra, err := iprange.New(allowRange)
	if err != nil {
		return err
	}

	// register service
	var server *grpc.Server

	ipInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			logger.Crit("error fetching ip addr from request", nil)
			return nil, fmt.Errorf("error fetching ip addr from request")
		}

		ipAndPort := p.Addr.String()
		ip := strings.Split(ipAndPort, ":")

		if !ra.IncludeStr(ip[0]) {
			logger.Crit("not in allow ip range", ip[0], ra)
			return nil, fmt.Errorf("not in allow ip range", ip[0])
		}

		// Calls the handler
		return handler(ctx, req)
	}

	if creds == nil {
		// insecure
		server = grpc.NewServer(grpc.UnaryInterceptor(ipInterceptor))
	} else {
		// secure
		server = grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(ipInterceptor))
	}

	pb.RegisterVimonadeServiceServer(server, srv)

	// start gRPC server
	logger.Info("starting gRPC server on " + serverAddr)

	return server.Serve(listen)
}
