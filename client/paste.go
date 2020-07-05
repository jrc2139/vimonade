package client

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

func Paste(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug("failed to dial server: " + err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	var text string

	text = lc.pasteText(isConnected)
	if _, err := c.Out.Write([]byte(text)); err != nil {
		logger.Error("Failed to output Paste to stdin: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) pasteText(cnx bool) string {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		c.logger.Error("error reading from clipboard: " + err.Error())
	}

	if cnx {
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		if _, err := c.grpcClient.Paste(ctx, &pb.PasteRequest{Value: text}); err != nil {
			c.logger.Debug("error with client pasting " + err.Error())
		}
	}

	return lemon.ConvertLineEnding(text, c.lineEnding)
}
