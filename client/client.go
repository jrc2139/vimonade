package client

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/golang/protobuf/ptypes/wrappers"
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

	// not interested in copying blank and newlines
	switch text {
	case "":
		return nil
	case "\n":
		return nil
	default:
		_, err := c.grpcClient.Copy(context.Background(), &wrappers.StringValue{Value: text})
		if err != nil {
			c.logger.Debug(err.Error())
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		return err
	}

	return nil
}

func (c *client) Paste() (string, error) {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}

	if _, err := c.grpcClient.Paste(context.Background(), &wrappers.StringValue{Value: text}); err != nil {
		c.logger.Debug(err.Error())
	}

	return lemon.ConvertLineEnding(text, c.lineEnding), nil
}
