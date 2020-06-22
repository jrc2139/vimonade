package client

import (
	"context"
	"fmt"
	"time"

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

func (c *client) copy(text string) error {
	c.logger.Debug("Sending: " + text)

	// not interested in copying blank and newlines
	switch text {
	case "":
		return nil
	case "\n":
		return nil
	default:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := c.grpcClient.Copy(ctx, &wrappers.StringValue{Value: text})
		if err != nil {
			c.logger.Debug(err.Error())
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		return err
	}

	return nil
}

func (c *client) paste() (string, error) {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.grpcClient.Paste(ctx, &wrappers.StringValue{Value: text}); err != nil {
		c.logger.Debug(err.Error())
	}

	return lemon.ConvertLineEnding(text, c.lineEnding), nil
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}

func Copy(c *lemon.CLI, logger log.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Debug(err.Error())
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.copy(c.DataSource); err != nil {
		logger.Crit("Failed to Copy", err, nil)
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func Paste(c *lemon.CLI, logger log.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug(err.Error())
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	var text string

	text, err = lc.paste()
	if err != nil {
		logger.Crit("Failed to Paste", err, nil)
		writeError(c, err)

		return lemon.RPCError
	}

	if _, err := c.Out.Write([]byte(text)); err != nil {
		logger.Crit("Failed to output Paste to stdin", err, nil)
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}
