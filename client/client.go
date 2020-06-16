package client

import (
	"context"
	"fmt"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"

	"github.com/jrc2139/vimonade/lemon"
	pb "github.com/jrc2139/vimonade/pkg/api/v1"
)

type client struct {
	host               string
	port               int
	addr               string
	lineEnding         string
	noFallbackMessages bool
	logger             log.Logger
	grpcClient         pb.MessageServiceClient
}

// func Orig(c *lemon.CLI, logger log.Logger) *client {
// return &client{
// host:               c.Host,
// port:               c.Port,
// addr:               fmt.Sprintf("http://%s:%d", c.Host, c.Port),
// lineEnding:         c.LineEnding,
// noFallbackMessages: c.NoFallbackMessages,
// logger:             logger,
// }
// }

func New(c *lemon.CLI, conn *grpc.ClientConn, logger log.Logger) *client {
	return &client{
		host:               c.Host,
		port:               c.Port,
		addr:               fmt.Sprintf("http://%s:%d", c.Host, c.Port),
		lineEnding:         c.LineEnding,
		noFallbackMessages: c.NoFallbackMessages,
		logger:             logger,
		grpcClient:         pb.NewMessageServiceClient(conn),
	}
}

const MSGPACK = "application/x-msgpack"

func (c *client) Copy(text string) error {
	c.logger.Debug("Sending: " + text)

	_, err := c.grpcClient.Copy(context.Background(), &pb.Message{Text: text})
	if err != nil {
		// clipboard.WriteAll(text)
		return err
	}

	return nil
}

func (c *client) Paste() (string, error) {
	c.logger.Debug("Receiving")

	text, err := c.grpcClient.Paste(context.Background(), nil)
	if err != nil {
		return "", err
	}

	return lemon.ConvertLineEnding(text.Text, c.lineEnding), nil
}
