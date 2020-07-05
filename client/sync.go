package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	gosync "github.com/Redundancy/go-sync"
	"github.com/Redundancy/go-sync/blocksources"
	"github.com/Redundancy/go-sync/filechecksum"
	"github.com/Redundancy/go-sync/indexbuilder"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

type pathData struct {
	path string
	info os.FileInfo
}

// TODO consider abs obsolete
type pathDetail struct {
	abs string
	rel string
	dir bool
}

type filePath struct {
	abs string
	rel string
}

type dirPath struct {
	abs string
	rel string
}

func Sync(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Error("failed to dial server: " + err.Error())
		return lemon.RPCError
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	// errChan := make(chan error)
	// doneChan := make(chan struct{})
	// defer close(doneChan)
	if err := lc.runRsyncPipeline(c.DataSource); err != nil {
		logger.Error("failed to syncFiles: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	// if err := lc.syncFiles(c.DataSource, errChan, doneChan); err != nil {
	// if err := lc.syncFiles(c.DataSource); err != nil {
	// logger.Error("failed to syncFiles: " + err.Error())
	// writeError(c, err)
	//
	// return lemon.RPCError
	// }

	// select {
	// case err := <-errChan:
	// logger.Error("failed to syncFiles: " + err.Error())
	// writeError(c, err)
	//
	// return lemon.RPCError
	// case <-doneChan:
	// return lemon.Success
	// }
	return lemon.Success
}

// func (c *client) syncFiles(dir string, errCh chan error, doneCh chan struct{}) error {
func (c *client) syncFiles(dir string) error {
	c.logger.Debug("Syncing directory: " + dir)
	var dirPath string

	if dir == "" {
		return nil
	}

	if dir == "." {
		absPath, err := filepath.Abs(filepath.Dir(dir))
		if err != nil {
			return err
		}

		dirPath = absPath
	} else {
		dirPath = dir
	}

	fmt.Println(dirPath)

	// check if dir is accessible
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}

	// fmt.Println("########", skipFileTypes, skipDirs)

	// filePathChan := c.readFiles(dirPath, skipFileTypes, skipDirs)

	// for _, f := range filePaths {
	// fmt.Printf("%+v\n", f)
	// }

	return nil
}

// TODO consider losing abs
// pd.rel is for server
// pd.abs is for local
func (c *client) syncFile(pd pathDetail) error {
	c.logger.Debug("Syncing file: " + pd.rel)

	if pd.dir {
	}

	remoteData, err := c.receiveFile(pd.abs)
	if err != nil {
		return err
	}

	data, err := c.performRsync(remoteData, pd.abs)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	stream, err := c.grpcClient.Send(ctx)
	if err != nil {
		return err
	}

	req := &pb.SendFileRequest{
		Data: &pb.SendFileRequest_Info{
			Info: &pb.FileInfo{
				Name:     filepath.Base(pd.rel),
				FileType: filepath.Ext(pd.rel),
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	for _, d := range data {
		req := &pb.SendFileRequest{
			Data: &pb.SendFileRequest_ChunkData{
				ChunkData: d,
			},
		}

		err = stream.Send(req)
		if err != nil {
			return err
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	c.logger.Debug(fmt.Sprintf("file sent with id: %s, size: %d", res.GetName(), res.GetSize()))

	return nil
}

func (c *client) performRsync(remoteData []byte, path string) ([][]byte, error) {
	localData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	generator := filechecksum.NewFileChecksumGenerator(BLOCK_SIZE)
	// _, referenceFileIndex, checksumLookup, err := indexbuilder.BuildIndexFromString(generator, REFERENCE)
	_, referenceFileIndex, checksumLookup, err := indexbuilder.BuildChecksumIndex(generator, file)
	if err != nil {
		return nil, err
	}

	fileSize := int64(len(localData))

	blockCount := fileSize / BLOCK_SIZE
	if fileSize%BLOCK_SIZE != 0 {
		blockCount++
	}

	fs := &gosync.BasicSummary{
		ChecksumIndex:  referenceFileIndex,
		ChecksumLookup: checksumLookup,
		BlockCount:     uint(blockCount),
		BlockSize:      uint(BLOCK_SIZE),
		FileSize:       fileSize,
	}

	c.logger.Debug(fmt.Sprintf("%+v", fs))

	inputFile := bytes.NewReader(remoteData)
	patchedFile := bytes.NewBuffer(nil)

	rsync := &gosync.RSync{
		Input:  inputFile,
		Output: patchedFile,
		Source: blocksources.NewReadSeekerBlockSource(
			bytes.NewReader(localData),
			blocksources.MakeNullFixedSizeResolver(uint64(BLOCK_SIZE)),
		),
		Summary: fs,
		OnClose: nil,
	}

	c.logger.Debug(fmt.Sprintf("####rsync\n%+v", rsync))

	// rsync, err := gosync.MakeRSync(
	// path,
	// referencePath,
	// outFilename,
	// fs,
	// )

	// if err != nil {
	// return nil, err
	// }

	if err := rsync.Patch(); err != nil {
		return nil, err
	}

	if err := rsync.Close(); err != nil {
		return nil, err
	}

	c.logger.Debug(fmt.Sprintf("##patchedFile\n%+v", patchedFile))

	return makeBuffer(patchedFile.Bytes()), nil
}

func (c *client) receiveFile(path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	req := &pb.SyncFileRequest{
		Name: filepath.Base(path),
	}

	stream, err := c.grpcClient.Sync(ctx, req)
	if err != nil {
		return nil, err
	}

	data := bytes.Buffer{}
	fSize := 0

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			c.logger.Debug("no more data")
			break
		}
		if err != nil {
			c.logger.Error(fmt.Sprintf("cannot receive chunk data: %v", err))
			return nil, err
		}

		chunk := res.GetChunkData()
		size := len(chunk)

		c.logger.Debug(fmt.Sprintf("received a chunk with size %d", size))

		fSize += size
		if fSize > maxFileSize {
			c.logger.Error(fmt.Sprintf("f is too large: %d > %d", fSize, maxFileSize))
			return nil, fmt.Errorf("max file size exceeded")
		}

		_, err = data.Write(chunk)
		if err != nil {
			c.logger.Error(fmt.Sprintf("f is too large: %d > %d", fSize, maxFileSize))
			return nil, err
		}
	}

	return data.Bytes(), nil
}

func readGitIgnore(path string) ([]string, []string, error) {
	dirs := []string{".git"}
	files := []string{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return files, dirs, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return files, dirs, err
	}

	defer func() {
		err = f.Close()
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "\n" {
			continue
		}
		line := strings.TrimSpace(s.Text())

		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "*.") {
			files = append(files, strings.Split(line, "*")[1])
			continue
		}

		dirs = append(dirs, line)
	}

	err = s.Err()
	if err != nil {
		return files, dirs, err
	}

	return dirs, files, nil
}

func (c *client) syncDirs(ctx context.Context, dirs <-chan dirPath) <-chan error {
	c.logger.Debug("Making remote directories")
	errc := make(chan error, 1)

	go func() {
		defer close(errc)

		clientCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := c.grpcClient.MakeDir(clientCtx)
		if err != nil {
			errc <- fmt.Errorf("cannot make remote directory %v", err)
			return
		}

		// waitResponse := make(chan error)

		// go routine to receive responses
		go func() {
			for {
				_, err := stream.Recv()
				if err == io.EOF {
					c.logger.Debug("no more responses")
					// waitResponse <- nil
					return
				}
				if err != nil {
					errc <- fmt.Errorf("cannot receive makedir stream response: %v", err)
					return
				}

				c.logger.Debug("received response")
			}
		}()

		for d := range dirs {
			req := &pb.DirRequest{Name: d.abs}
			err := stream.Send(req)
			if err != nil {
				errc <- fmt.Errorf("cannot send makedir request: %v - %v", err, stream.RecvMsg(nil))
				// return errc
				break
			}

			c.logger.Debug("received response")

			err = stream.CloseSend()
			if err != nil {
				errc <- fmt.Errorf("cannot close send: %v", err)
				break
				// return errc
			}

			// err = <-waitResponse

			// return errc
		}
	}()

	return errc
}
