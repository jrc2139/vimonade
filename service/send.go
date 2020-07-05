package service

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/jrc2139/vimonade/api"
)

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
