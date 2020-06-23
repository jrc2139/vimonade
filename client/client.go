package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

const (
	timeOut = 5 * time.Second
)

type client struct {
	host       string
	port       int
	lineEnding string
	logger     *zap.Logger
	grpcClient pb.VimonadeServiceClient
}

func New(c *lemon.CLI, conn *grpc.ClientConn, logger *zap.Logger) *client {
	return &client{
		host:       c.Host,
		port:       c.Port,
		lineEnding: c.LineEnding,
		logger:     logger,
		grpcClient: pb.NewVimonadeServiceClient(conn),
	}
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

			_, err := c.grpcClient.Copy(ctx, &wrappers.StringValue{Value: text})
			if err != nil {
				c.logger.Debug("error with client copying " + err.Error())
			}
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		return err
	}

	return nil
}

func (c *client) pasteText(cnx bool) (string, error) {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}

	if cnx {
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		if _, err := c.grpcClient.Paste(ctx, &wrappers.StringValue{Value: text}); err != nil {
			c.logger.Debug("error with client pasting " + err.Error())
		}
	}

	return lemon.ConvertLineEnding(text, c.lineEnding), nil
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}

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

	text, err = lc.pasteText(isConnected)
	if err != nil {
		logger.Error("Failed to Paste: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	if _, err := c.Out.Write([]byte(text)); err != nil {
		logger.Error("Failed to output Paste to stdin: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func Send(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Fatal("failed to dial server: " + err.Error())
		return lemon.RPCError
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.send(c.DataSource); err != nil {
		logger.Fatal("failed to send: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) send(path string) error {
	c.logger.Debug("Sending " + path)

	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	stream, err := c.grpcClient.Send(ctx)
	if err != nil {
		return err
	}

	req := &pb.SendFileRequest{
		Data: &pb.SendFileRequest_Info{
			Info: &pb.FileInfo{
				Name:     filepath.Base(path),
				FileType: filepath.Ext(path),
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		req := &pb.SendFileRequest{
			Data: &pb.SendFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			return err
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	c.logger.Debug(fmt.Sprintf("image sent with id: %s, size: %d", res.GetName(), res.GetSize()))

	return nil
}
