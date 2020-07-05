package service

import (
	"context"
	"fmt"

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
	fileStore   FileStore
	lineEnding  string
	vimonadeDir string
	// path       string
	logger *zap.Logger
}

// NewVimonadeServerService creates Audio service object.
func NewVimonadeServerService(fileStore FileStore, vimonadeDir, lineEnding string, logger *zap.Logger) pb.VimonadeServiceServer {
	return &vimonadeServiceServer{
		fileStore:   fileStore,
		vimonadeDir: vimonadeDir,
		lineEnding:  lineEnding,
		logger:      logger,
	}
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
