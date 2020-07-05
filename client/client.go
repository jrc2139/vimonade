package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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

type client struct {
	host       string
	port       int
	lineEnding string
	logger     *zap.Logger
	grpcClient pb.VimonadeServiceClient
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

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
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
