package service

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/jrc2139/vimonade/api"
)

func (s *vimonadeServiceServer) MakeDir(stream pb.VimonadeService_MakeDirServer) error {
	for {
		err := s.contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			s.logger.Debug("makedir stream ended")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive stream request: %v", err))
		}

		if err := s.createDir(req.GetName()); err != nil {
			s.logger.Error(fmt.Sprintf("error creating dir %s: %s", req.GetName(), err))

			return logError(status.Errorf(codes.Unknown, "error creating dir: %v", err))
		}

		res := &pb.DirResponse{}

		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot send stream response: %v", err))
		}
	}

	return nil
}

func (s *vimonadeServiceServer) createDir(name string) error {
	path := s.vimonadeDir + "/" + name
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
