package service

import (
	"context"

	"github.com/atotto/clipboard"

	pb "github.com/jrc2139/vimonade/api"
)

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
