package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/atotto/clipboard"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/jrc2139/vimonade/api"
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
	logger *zap.Logger
}

// NewVimonadeServerService creates Audio service object.
func NewVimonadeServerService(fileStore FileStore, lineEnding string, logger *zap.Logger) pb.VimonadeServiceServer {
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

func (s *vimonadeServiceServer) Copy(ctx context.Context, message *pb.CopyRequest) (*pb.CopyResponse, error) {
	err := s.contextError(ctx)
	if err != nil {
		return &pb.CopyResponse{}, err
	}

	if message != nil {
		s.logger.Debug("Copy requested: message: " + message.GetValue())

		if err := clipboard.WriteAll(message.GetValue()); err != nil {
			s.logger.Fatal("Writing to clipboard failed: " + err.Error())
			return &pb.CopyResponse{}, err
		}
	} else {
		s.logger.Debug("Copy requested: message=<empty>")
	}

	return &pb.CopyResponse{}, err
}

func (s *vimonadeServiceServer) Paste(ctx context.Context, message *pb.PasteRequest) (*pb.PasteResponse, error) {
	err := s.contextError(ctx)
	if err != nil {
		return &pb.PasteResponse{}, err
	}

	if message != nil {
		s.logger.Debug("Paste requested: message: " + message.GetValue())

		_, err := clipboard.ReadAll()
		if err != nil {
			s.logger.Fatal("Reading from clipboard failed: " + err.Error())
			return &pb.PasteResponse{}, err
		}
	} else {
		s.logger.Debug("Paste requested: message=<empty>")
	}

	return &pb.PasteResponse{}, nil
}

func logError(err error) error {
	if err != nil {
		fmt.Print(err)
	}

	return err
}

func (s *vimonadeServiceServer) contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		s.logger.Debug("request is canceled")
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		s.logger.Debug("deadline is exceeded")
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}
