package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/atotto/clipboard"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

const (
	maxFileSize = 1 << 30
)

// VimonadeServer is implementation of pb.VimonadeServer proto interface.
type vimonadeServiceServer struct {
	// localStore LocalStore
	fileStore  FileStore
	lineEnding string
	// path       string
	logger log.Logger
}

// NewVimonadeServerService creates Audio service object.
func NewVimonadeServerService(fileStore FileStore, lineEnding string, logger log.Logger) pb.VimonadeServiceServer {
	return &vimonadeServiceServer{fileStore: fileStore, lineEnding: lineEnding, logger: logger}
}

func (s *vimonadeServiceServer) Send(stream pb.VimonadeService_SendServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive file info"))
	}

	name := req.GetInfo().GetName()
	fileType := req.GetInfo().GetFileType()

	s.logger.Info("receive an send-file request for " + name)

	// laptop, err := server.laptopStore.Find(id)
	// if err != nil {
	// return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
	// }
	// if laptop == nil {
	// return logError(status.Errorf(codes.InvalidArgument, "laptop id %s
	// doesn't exist", id))
	// }

	fData := bytes.Buffer{}
	fSize := 0

	for {
		s.logger.Debug("waiting to receive more data")

		err := s.contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			s.logger.Debug("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		s.logger.Debug(fmt.Sprintf("received a chunk with size %d", size))

		fSize += size
		if fSize > maxFileSize {
			return logError(status.Errorf(codes.InvalidArgument, "f is too large: %d > %d", fSize, maxFileSize))
		}
		_, err = fData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	savedName, err := s.fileStore.Save(name, fileType, fData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save file to the store: %v", err))
	}

	res := &pb.SendFileResponse{
		Name: savedName,
		Size: uint32(fSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	s.logger.Debug(fmt.Sprintf("saved file %s with size %d", name, fSize))

	return nil
}

func (s *vimonadeServiceServer) Copy(ctx context.Context, message *wrappers.StringValue) (*empty.Empty, error) {
	err := s.contextError(ctx)
	if err != nil {
		return &empty.Empty{}, err
	}

	if message != nil {
		s.logger.Debug("Copy requested: message: " + message.GetValue())

		if err := clipboard.WriteAll(message.GetValue()); err != nil {
			s.logger.Debug("Writing to clipboard failed: %v", err)
			return &empty.Empty{}, err
		}
	} else {
		s.logger.Debug("Copy requested: message=<empty>")
	}

	return &empty.Empty{}, nil
}

func (s *vimonadeServiceServer) Paste(ctx context.Context, message *wrappers.StringValue) (*wrappers.StringValue, error) {
	err := s.contextError(ctx)
	if err != nil {
		return &wrappers.StringValue{Value: ""}, err
	}

	if message != nil {
		s.logger.Debug("Paste requested: message: " + message.GetValue())

		_, err := clipboard.ReadAll()
		if err != nil {
			s.logger.Debug("Writing to clipboard failed: %v", err)
			return &wrappers.StringValue{Value: ""}, err
		}
	} else {
		s.logger.Debug("Paste requested: message=<empty>")
	}

	return &wrappers.StringValue{Value: lemon.ConvertLineEnding(message.GetValue(), s.lineEnding)}, nil
}

func logError(err error) error {
	if err != nil {
		fmt.Print(err)
	}

	return err
}

/*
import (
	"context"
	"fmt"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc/credentials"

	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/protocol/grpc"
	"github.com/jrc2139/vimonade/service"
)

func Serve(c *lemon.CLI, creds credentials.TransportCredentials, logger log.Logger) error {
	// Server
	if err := grpc.RunServer(context.Background(), service.NewVimonadeServerService(c.LineEnding, logger), creds, c.Allow, fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		return err
	}

	return nil
}
*/

func (s *vimonadeServiceServer) contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		s.logger.Debug("request is canceled")
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		s.logger.Debug("request is canceled")
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}
