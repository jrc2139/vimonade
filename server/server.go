package server

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

	"github.com/jrc2139/vimonade/lemon"
	v1 "github.com/jrc2139/vimonade/pkg/api/v1"
	service "github.com/jrc2139/vimonade/pkg/service/v1"
)

func Serve(c *lemon.CLI, creds credentials.TransportCredentials, logger log.Logger) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}

	ra, err := iprange.New(c.Allow)
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

	srv := service.NewMessageServerService(c.LineEnding, logger)

	v1.RegisterMessageServiceServer(server, srv)

	// start gRPC server
	logger.Debug("starting gRPC server...")

	return server.Serve(listen)
}
