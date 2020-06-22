package v1

import (
	"context"
	"errors"

	"github.com/atotto/clipboard"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"

	log "github.com/inconshreveable/log15"

	"github.com/jrc2139/vimonade/lemon"
	v1 "github.com/jrc2139/vimonade/pkg/api/v1"
)

// MessageServer is implementation of v1.MessageServer proto interface.
type messageServiceServer struct {
	lineEnding string
	logger     log.Logger
}

// NewMessageServerService creates Audio service object.
func NewMessageServerService(lineEnding string, logger log.Logger) v1.MessageServiceServer {
	return &messageServiceServer{lineEnding: lineEnding, logger: logger}
}

func (s *messageServiceServer) Copy(ctx context.Context, message *wrappers.StringValue) (*empty.Empty, error) {
	if errors.Is(ctx.Err(), context.Canceled) {
		return &empty.Empty{}, ctx.Err()
	}

	if message != nil {
		s.logger.Debug("Copy requested: message=%v", message)

		if err := clipboard.WriteAll(message.Value); err != nil {
			s.logger.Debug("Writing to clipboard failed: %v", err)
			return &empty.Empty{}, err
		}
	} else {
		s.logger.Debug("Copy requested: message=<empty>")
	}

	return &empty.Empty{}, nil
}

func (s *messageServiceServer) Paste(ctx context.Context, message *wrappers.StringValue) (*wrappers.StringValue, error) {
	if errors.Is(ctx.Err(), context.Canceled) {
		return &wrappers.StringValue{Value: ""}, ctx.Err()
	}

	if message != nil {
		s.logger.Debug("Paste requested: message=%v", message)

		_, err := clipboard.ReadAll()
		if err != nil {
			s.logger.Debug("Writing to clipboard failed: %v", err)
			return &wrappers.StringValue{Value: ""}, err
		}
	} else {
		s.logger.Debug("Paste requested: message=<empty>")
	}

	return &wrappers.StringValue{Value: lemon.ConvertLineEnding(message.Value, s.lineEnding)}, nil
}
