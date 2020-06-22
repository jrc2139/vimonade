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
	log "github.com/inconshreveable/log15"
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
	logger     log.Logger
	grpcClient pb.VimonadeServiceClient
}

func New(c *lemon.CLI, conn *grpc.ClientConn, logger log.Logger) *client {
	return &client{
		host:       c.Host,
		port:       c.Port,
		lineEnding: c.LineEnding,
		logger:     logger,
		grpcClient: pb.NewVimonadeServiceClient(conn),
	}
}

func (c *client) copy(text string, cnx bool) error {
	c.logger.Debug("Sending: " + text)

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
				c.logger.Debug(err.Error())
			}
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		return err
	}

	return nil
}

func (c *client) paste(cnx bool) (string, error) {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}

	if cnx {
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		if _, err := c.grpcClient.Paste(ctx, &wrappers.StringValue{Value: text}); err != nil {
			c.logger.Debug(err.Error())
		}
	}

	return lemon.ConvertLineEnding(text, c.lineEnding), nil
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}

func Copy(c *lemon.CLI, logger log.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug(err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.copy(c.DataSource, isConnected); err != nil {
		logger.Crit("Failed to Copy", err, nil)
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func Paste(c *lemon.CLI, logger log.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug(err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	var text string

	text, err = lc.paste(isConnected)
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

func Send(c *lemon.CLI, logger log.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Crit(err.Error())
		return lemon.RPCError
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.send(c.DataSource); err != nil {
		logger.Crit("Failed to Copy", err, nil)
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
		c.logger.Crit("cannot send file", err)
	}

	req := &pb.SendFileRequest{
		Data: &pb.SendFileRequest_Info{
			Info: &pb.FileInfo{
				Name:     filepath.Base(path),
				FileType: filepath.Ext(path),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		c.logger.Crit("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			c.logger.Crit("cannot read chunk to buffer: ", err)
		}

		req := &pb.SendFileRequest{
			Data: &pb.SendFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			c.logger.Crit("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		c.logger.Crit("cannot receive response: ", err)
	}

	c.logger.Debug(fmt.Sprintf("image sent with id: %s, size: %d", res.GetName(), res.GetSize()))

	return nil
}
