package service

import (
	"context"

	"github.com/atotto/clipboard"

	pb "github.com/jrc2139/vimonade/api"
)

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
