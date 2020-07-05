package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

func Copy(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug("failed to dial server: " + err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.copyText(c.DataSource, isConnected); err != nil {
		logger.Error("failed to Copy: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) copyText(text string, cnx bool) error {
	c.logger.Debug("Copying: " + text)

	// not interested in copying blank and newlines
	switch text {
	case "":
		return nil
	case "\n":
		return nil
	default:
		if cnx {
			ctx, cancel := context.WithTimeout(context.Background(), timeOut)
			defer cancel()

			_, err := c.grpcClient.Copy(ctx, &pb.CopyRequest{Value: strings.TrimSpace(text)})
			if err != nil {
				c.logger.Debug("error with client copying " + err.Error())
			}
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		c.logger.Error("error writing to clipboard: " + err.Error())
	}

	return nil
}
