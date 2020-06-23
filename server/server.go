package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pocke/go-iprange"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/service"
)

func Serve(c *lemon.CLI, creds credentials.TransportCredentials, logger *zap.Logger) int {
	// create vimonade dir if !exist
	var vimonadeDir string

	if c.VimonadeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Cannot find $HOME error: " + err.Error())
			return lemon.RPCError
		}

		vimonadeDir = home + "/.vimonade/files"
	} else {
		vimonadeDir = c.VimonadeDir
	}

	logger.Debug("current vimonade dir: " + vimonadeDir)

	if _, err := os.Stat(vimonadeDir); os.IsNotExist(err) {
		err = os.MkdirAll(vimonadeDir, 0755)
		if err != nil {
			logger.Error("Creating vimonade dir error: " + err.Error())
			return lemon.RPCError
		}
	}

	// Server
	store := service.NewDiskFileStore(vimonadeDir)

	if err := runServer(context.Background(),
		service.NewVimonadeServerService(store, c.LineEnding, logger),
		logger, creds, c.Allow, fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		logger.Error("Server error: " + err.Error())
		fmt.Fprintln(c.Err, err.Error())

		return lemon.RPCError
	}

	return lemon.RPCError
}

// runServer registers gRPC service and run server.
func runServer(ctx context.Context, srv pb.VimonadeServiceServer, logger *zap.Logger, creds credentials.TransportCredentials, allowRange, serverAddr string) error {
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
			logger.Error("error fetching ip addr from request")
			return nil, fmt.Errorf("error fetching ip addr from request")
		}

		ipAndPort := p.Addr.String()
		ip := strings.Split(ipAndPort, ":")

		if !ra.IncludeStr(ip[0]) {
			logger.Error(fmt.Sprintf("not in allow ip range: %s | %s", ip[0], ra))
			return nil, fmt.Errorf("not in allow ip range: %s", ip[0])
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
