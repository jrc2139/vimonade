package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/jrc2139/vimonade/api"
)

func (s *vimonadeServiceServer) Sync(req *pb.SyncFileRequest, stream pb.VimonadeService_SyncServer) error {
	name := req.GetName()
	if name == "" {
		return nil
	}

	path := s.vimonadeDir + "/" + name

	s.logger.Info("receive an sync-file request for " + path)

	// don't continue if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	resp := &pb.SyncFileResponse{
		Data: &pb.SyncFileResponse_Info{
			Info: &pb.FileInfo{
				Name:     path,
				FileType: filepath.Ext(path),
			},
		},
	}

	if err := stream.Send(resp); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		err := s.contextError(stream.Context())
		if err != nil {
			return err
		}

		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		req := &pb.SyncFileResponse{
			Data: &pb.SyncFileResponse_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			return err
		}
	}

	s.logger.Debug(fmt.Sprintf("file synced: %s", path))

	return nil
}
