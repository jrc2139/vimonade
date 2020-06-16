package v1

import (
	"context"
	"log"

	"github.com/atotto/clipboard"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/jrc2139/vimonade/lemon"
	v1 "github.com/jrc2139/vimonade/pkg/api/v1"
)

// MessageServer is implementation of v1.MessageServer proto interface
type messageServiceServer struct {
	lineEnding string
}

// NewMessageServerService creates Audio service object
func NewMessageServerService(lineEnding string) v1.MessageServiceServer {
	return &messageServiceServer{lineEnding: lineEnding}
}

func (s *messageServiceServer) Copy(ctx context.Context, message *v1.Message) (*empty.Empty, error) {
	if message != nil {
		log.Printf("Copy requested: message=%v", *message)

		if err := clipboard.WriteAll(message.Text); err != nil {
			log.Printf("Writing to clipboard failed: %v", err)
			return &empty.Empty{}, err
		}
	} else {
		log.Print("Copy requested: message=<empty>")
	}

	return &empty.Empty{}, nil
}

func (s *messageServiceServer) Paste(ctx context.Context, empty *empty.Empty) (*v1.Message, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Error reading from clipboard=%v", err)

		return &v1.Message{}, nil
	}

	return &v1.Message{Text: lemon.ConvertLineEnding(text, s.lineEnding)}, nil
}
