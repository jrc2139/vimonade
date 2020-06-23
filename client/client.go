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
	"sync"
	"time"

	gosync "github.com/Redundancy/go-sync"
	"github.com/Redundancy/go-sync/blocksources"
	"github.com/Redundancy/go-sync/filechecksum"
	"github.com/Redundancy/go-sync/indexbuilder"
	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/jrc2139/vimonade/api"
	"github.com/jrc2139/vimonade/lemon"
)

const (
	timeOut     = 5 * time.Second
	BLOCK_SIZE  = 128
	maxFileSize = 1 << 30
)

/* type filePath struct { */
// abs string
// rel string
// dir bool
/* } */

type client struct {
	host       string
	port       int
	lineEnding string
	logger     *zap.Logger
	grpcClient pb.VimonadeServiceClient
}

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

func New(c *lemon.CLI, conn *grpc.ClientConn, logger *zap.Logger) *client {
	return &client{
		host:       c.Host,
		port:       c.Port,
		lineEnding: c.LineEnding,
		logger:     logger,
		grpcClient: pb.NewVimonadeServiceClient(conn),
	}
}
func Copy(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug("failed to dial server: " + err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.copyText(c.DataSource, isConnected); err != nil {
		logger.Error("failed to Copy: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) copyText(text string, cnx bool) error {
	c.logger.Debug("Copying: " + text)

	// not interested in copying blank and newlines
	switch text {
	case "":
		return nil
	case "\n":
		return nil
	default:
		if cnx {
			ctx, cancel := context.WithTimeout(context.Background(), timeOut)
			defer cancel()

			_, err := c.grpcClient.Copy(ctx, &pb.CopyRequest{Value: strings.TrimSpace(text)})
			if err != nil {
				c.logger.Debug("error with client copying " + err.Error())
			}
		}
	}

	if err := clipboard.WriteAll(text); err != nil {
		c.logger.Error("error writing to clipboard: " + err.Error())
	}

	return nil
}

func Paste(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	isConnected := true

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		// don't return err if connection isn't made
		logger.Debug("failed to dial server: " + err.Error())
		isConnected = false
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	var text string

	text = lc.pasteText(isConnected)
	if _, err := c.Out.Write([]byte(text)); err != nil {
		logger.Error("Failed to output Paste to stdin: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
}

func (c *client) pasteText(cnx bool) string {
	c.logger.Debug("Receiving")

	text, err := clipboard.ReadAll()
	if err != nil {
		c.logger.Error("error reading from clipboard: " + err.Error())
	}

	if cnx {
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		if _, err := c.grpcClient.Paste(ctx, &pb.PasteRequest{Value: text}); err != nil {
			c.logger.Debug("error with client pasting " + err.Error())
		}
	}

	return lemon.ConvertLineEnding(text, c.lineEnding)
}

func Send(c *lemon.CLI, logger *zap.Logger, opts ...grpc.DialOption) int {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), opts...)
	if err != nil {
		logger.Fatal("failed to dial server: " + err.Error())
		return lemon.RPCError
	}
	defer conn.Close()

	lc := New(c, conn, logger)

	if err := lc.sendFile(c.DataSource); err != nil {
		logger.Fatal("failed to sendFile: " + err.Error())
		writeError(c, err)

		return lemon.RPCError
	}

	return lemon.Success
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

func (c *client) sendFile(path string) error {
	c.logger.Debug("Sending " + path)

	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	stream, err := c.grpcClient.Send(ctx)
	if err != nil {
		return err
	}

	req := &pb.SendFileRequest{
		Data: &pb.SendFileRequest_Info{
			Info: &pb.FileInfo{
				Name:     filepath.Base(path),
				FileType: filepath.Ext(path),
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		req := &pb.SendFileRequest{
			Data: &pb.SendFileRequest_ChunkData{
				ChunkData: buffer[:n],
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

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
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

func makeBuffer(data []byte) [][]byte {
	bufferSize := 1024
	buffer := make([][]byte, 0, (len(data)+bufferSize-1)/bufferSize)

	for bufferSize < len(data) {
		data, buffer = data[bufferSize:], append(buffer, data[0:bufferSize:bufferSize])
	}
	buffer = append(buffer, data)

	return buffer
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

func (c *client) readFiles(dirPath string, skipFileTypes, skipDirs []string) <-chan filePath {
	ch := make(chan filePath)

	go func() {
		// subDirToSkip := []string{".git"}
		err := filepath.Walk(dirPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					c.logger.Error(fmt.Sprintf("prevent panic by handling failure accessing a path %q: %v\n", path, err))
					return err
				}

				c.logger.Debug(fmt.Sprintf("visited file or dir: %s\n", path))

				// check if dir should be ignored
				if info.IsDir() {
					for _, s := range skipDirs {
						if info.Name() == s {
							c.logger.Info(info.Name())
							c.logger.Error(fmt.Sprintf("skipping a dir without errors: %+v \n", info.Name()))
							return filepath.SkipDir
						}
					}

					rel, err := filepath.Rel(dirPath, path)
					if err != nil {
						return err
					}

					ch <- filePath{abs: path, rel: rel}

					return nil
				}

				// check if file should be ignored
				ext := filepath.Ext(path)
				if ext != "" {
					for _, s := range skipFileTypes {
						if ext == s {
							return nil
						}

						if info.Name() == s {
							return nil
						}
					}

					rel, err := filepath.Rel(dirPath, path)
					if err != nil {
						return err
					}

					ch <- filePath{abs: path, rel: rel}

					return nil
				}

				// fmt.Println("####path", path)
				// fmt.Println(filepath.Clean(path))

				// if info.IsDir() && info.Name() == subDirToSkip {
				// c.logger.Debug(fmt.Sprintf("skipping a dir without errors: %+v \n", info.Name()))
				// return filepath.SkipDir
				// }

				// TODO logic for folder creation
				// if !info.IsDir() {
				// go func() {
				// if err := c.syncFile(path); err != nil {
				// return err
				// errCh <- err
				// }
				// }()
				// }
				// if err := c.syncFile(path); err != nil {
				// return err
				// }
				return nil
			}) //;err != nil {
		// return err
		if err != nil {
			c.logger.Error("file walk error " + err.Error())
		}

		close(ch)
	}()

	return ch
}

// func (c *client) sync(in <-chan filePath) <-chan error {
// out := make(chan error)
// go func() {
// for i := range in {
// out <- c.syncFile(i)
// }
// }()
// return out
// }

func indexFiles(ctx context.Context, dirPath string, skipFileTypes, skipDirs []string) (<-chan pathData, <-chan error) {
	out := make(chan pathData)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)
		// subDirToSkip := []string{".git"}
		err := filepath.Walk(dirPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					// c.logger.Error(fmt.Sprintf("prevent panic by handling failure accessing a path %q: %v\n", path, err))

					errc <- errors.Errorf("error accessing path: %s\n$s", path, err)
					return err
				}

				// c.logger.Debug(fmt.Sprintf("visited file or dir: %s\n", path))

				// check if dir should be ignored
				if info.IsDir() {
					for _, s := range skipDirs {
						if info.Name() == s {
							// c.logger.Info(info.Name())
							// c.logger.Error(fmt.Sprintf("skipping a dir without errors: %+v \n", info.Name()))
							return filepath.SkipDir
						}
					}

					// rel, err := filepath.Rel(dirPath, path)
					// if err != nil {
					// errc <- errors.Errorf("error finding rel path: %s", err)
					// return err
					// }

					// Send the data to the output channel but return early
					// if the context has been cancelled.
					select {
					case out <- pathData{path: path, info: info}:
					case <-ctx.Done():
						return nil
					}
					return nil
				}

				// check if file should be ignored
				ext := filepath.Ext(path)
				if ext != "" {
					for _, s := range skipFileTypes {
						if ext == s {
							return nil
						}

						if info.Name() == s {
							return nil
						}
					}

					// rel, err := filepath.Rel(dirPath, path)
					// if err != nil {
					// errc <- errors.Errorf("error finding rel path: %s", err)
					// return err
					// }

					select {
					case out <- pathData{path: path, info: info}:
					case <-ctx.Done():
						return nil
					}

					return nil
				}

				// fmt.Println("####path", path)
				// fmt.Println(filepath.Clean(path))

				// if info.IsDir() && info.Name() == subDirToSkip {
				// c.logger.Debug(fmt.Sprintf("skipping a dir without errors: %+v \n", info.Name()))
				// return filepath.SkipDir
				// }

				// TODO logic for folder creation
				// if !info.IsDir() {
				// go func() {
				// if err := c.syncFile(path); err != nil {
				// return err
				// errCh <- err
				// }
				// }()
				// }
				// if err := c.syncFile(path); err != nil {
				// return err
				// }
				return nil
			}) //;err != nil {
		// return err
		if err != nil {
			// c.logger.Error("file walk error " + err.Error())
			errc <- errors.Errorf("file walk error: %s", err)
		}
	}()

	return out, errc
}

func transformPathData(ctx context.Context, dirPath string, in <-chan pathData) (<-chan pathDetail, <-chan error) {
	out := make(chan pathDetail)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)
		for f := range in {
			rel, err := filepath.Rel(dirPath, f.path)
			if err != nil {
				errc <- errors.Errorf("error finding rel path: %s", err)
				return
			}

			select {
			case out <- pathDetail{abs: f.path, rel: rel, dir: f.info.IsDir()}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, errc
}

type filePath struct {
	abs string
	rel string
}

type dirPath struct {
	abs string
	rel string
}

func separateFilesAndDirs(ctx context.Context, in <-chan pathDetail) (<-chan dirPath, <-chan filePath) {
	fpOut := make(chan filePath)
	dpOut := make(chan dirPath)
	// done := make(chan struct{})

	go func() {
		defer close(fpOut)
		defer close(dpOut)
		// defer close(done)
		for p := range in {
			select {
			case <-ctx.Done():
				return
			default:
				if p.dir {
					dpOut <- dirPath{abs: p.abs, rel: p.rel}
				} else {
					fpOut <- filePath{abs: p.abs, rel: p.rel}
				}
			}
		}

		// done <- struct{}{}
	}()

	// <-done

	return dpOut, fpOut
}

// func (c *client) rsyncSink(ctx context.Context, in <-chan filePath) <-chan error {
// errc := make(chan error, 1)
//
// go func() {
// defer close(errc)
// for i := range in {
// if err := c.syncFile(i); err != nil {
// errc <- err
// return
// }
// }
// }()
//
// return errc
// }

func (c *client) runRsyncPipeline(dirPath string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// TODO add default dir to skip to config / lemon cli flag, take from
	// .gitignore
	// check if dir is accessible
	// var skip []string
	skipDirs, skipFileTypes, err := readGitIgnore(dirPath + "/.gitignore")
	if err != nil {
		return err
	}

	var errcList []<-chan error

	// Source pipeline stage.
	pathdatac, errc := indexFiles(ctx, dirPath, skipDirs, skipFileTypes)

	errcList = append(errcList, errc)

	// Transformer pipeline stage.
	pathdetailc, errc := transformPathData(ctx, dirPath, pathdatac)

	errcList = append(errcList, errc)

	dirpathc, _ := separateFilesAndDirs(ctx, pathdetailc)

	errc = c.syncDirs(ctx, dirpathc)

	errcList = append(errcList, errc)

	// Sink pipeline stage.
	// errc = c.rsyncSink(ctx, filepathc)
	// if err != nil {
	// return err
	// }

	// errcList = append(errcList, errc)

	fmt.Println("Pipeline started. Waiting for pipeline to complete.")

	return WaitForPipeline(errcList...)
}

// WaitForPipeline waits for results from all error channels.
// It returns early on the first error.
func WaitForPipeline(errs ...<-chan error) error {
	errc := MergeErrors(errs...)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

// MergeErrors merges multiple channels of errors.
// Based on https://blog.golang.org/pipelines.
func MergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	// We must ensure that the output channel has the capacity to
	// hold as many errors
	// as there are error channels.
	// This will ensure that it never blocks, even
	// if WaitForPipeline returns early.
	out := make(chan error, len(cs)) // Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls
	// wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	// Start a goroutine to close out once all the output goroutines
	// are done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
