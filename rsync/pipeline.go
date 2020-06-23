package rsync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

type pathData struct {
	path string
	info os.FileInfo
}

type pathDetail struct {
	abs string
	rel string
	dir bool
}

func indexFiles(ctx context.Context, dirPath string, skipFileTypes, skipDirs []string) (<-chan pathData, <-chan error, error) {
	out := make(chan pathData)
	errc := make(chan error, 1)

	go func() {
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

	return out, errc, nil
}

func transformPathData(ctx context.Context, dirPath string, in <-chan pathData) (<-chan pathDetail, <-chan error, error) {
	out := make(chan pathDetail)
	errc := make(chan error, 1)

	go func() {
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

	return out, errc, nil
}

func rsyncSink(ctx context.Context, dirPath string, in <-chan pathDetail) (<-chan error, error) {
	out := make(chan error, 1)

	go func() {
		// for i := range in {
		// out <- syncFile(i)
		// }
	}()

	return out, nil
}

func runRsyncPipeline(base int, lines []string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var errcList []<-chan error

	// Source pipeline stage.
	_, errc, err := indexFiles(ctx, ".", []string{}, []string{})
	if err != nil {
		return err
	}

	errcList = append(errcList, errc)

	// Transformer pipeline stage.
	// numberc, errc, err := lineParser(ctx, base, linec)
	// if err != nil {
	// return err
	// }

	errcList = append(errcList, errc)

	// Sink pipeline stage.
	// errc, err = sink(ctx, numberc)
	// if err != nil {
	// return err
	// }

	errcList = append(errcList, errc)

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

func lineListSource(ctx context.Context, lines ...string) (
	<-chan string, <-chan error, error) {
	if len(lines) == 0 {
		// Handle an error that occurs before the goroutine begins.
		return nil, nil, errors.Errorf("no lines provided")
	}

	out := make(chan string)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)
		for lineIndex, line := range lines {
			if line == "" {
				// Handle an error that occurs during the goroutine.
				errc <- errors.Errorf("line %v is empty", lineIndex+1)
				return
			}
			// Send the data to the output channel but return early
			// if the context has been cancelled.
			select {
			case out <- line:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}

func lineParser(ctx context.Context, base int, in <-chan string) (
	<-chan int64, <-chan error, error) {
	if base < 2 {
		// Handle an error that occurs before the goroutine begins.
		return nil, nil, errors.Errorf("invalid base %v", base)
	}
	out := make(chan int64)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errc)
		for line := range in {
			n, err := strconv.ParseInt(line, base, 64)
			if err != nil {
				// Handle an error that occurs during the goroutine.
				errc <- err
				return
			}
			// Send the data to the output channel but return early
			// if the context has been cancelled.
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}

func sink(ctx context.Context, in <-chan int64) (
	<-chan error, error) {
	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		for n := range in {
			if n >= 100 {
				// Handle an error that occurs during the goroutine.
				errc <- errors.Errorf("number %v is too large", n)
				return
			}
			fmt.Printf("sink: %v\n", n)
		}
	}()
	return errc, nil
}

func runSimplePipeline(base int, lines []string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var errcList []<-chan error

	// Source pipeline stage.
	linec, errc, err := lineListSource(ctx, lines...)
	if err != nil {
		return err
	}

	errcList = append(errcList, errc)

	// Transformer pipeline stage.
	numberc, errc, err := lineParser(ctx, base, linec)
	if err != nil {
		return err
	}

	errcList = append(errcList, errc)

	// Sink pipeline stage.
	errc, err = sink(ctx, numberc)
	if err != nil {
		return err
	}

	errcList = append(errcList, errc)

	fmt.Println("Pipeline started. Waiting for pipeline to complete.")

	return WaitForPipeline(errcList...)
}
