package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

func Send(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Fatal("failed to dial server: " + err.Error())
		return lemon.RPCError
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.sendFile(c.DataSource); err != nil {
		logger.Fatal("failed to sendFile: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) sendFile(path string) error {
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

	c.logger.Debug(fmt.Sprintf("file sent with id: %s, size: %d", res.GetName(), res.GetSize()))

	return nil
}
