package server

import (
	"context"
	"fmt"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc/credentials"

	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/pkg/protocol/grpc"
	v1 "github.com/jrc2139/vimonade/pkg/service/v1"
)

func Serve(c *lemon.CLI, creds credentials.TransportCredentials, logger log.Logger) error {
	// Server
	if err := grpc.RunServer(context.Background(), v1.NewMessageServerService(c.LineEnding, logger), creds, c.Allow, fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		return err
	}

	return nil
}
