package rsync

import (
	"time"

	"github.com/Redundancy/go-sync/blocksources"

	pb "github.com/jrc2139/vimonade/api"
)

const (
	timeOut = 5 * time.Second
	MB      = 1024 * 1024
)

func NewGrpcBlockSource(
	url string,
	concurrentRequests int,
	resolver blocksources.BlockSourceOffsetResolver,
	verifier blocksources.BlockVerifier,
	client pb.VimonadeServiceClient,
) *blocksources.BlockSourceBase {
	return blocksources.NewBlockSourceBase(
		&GrpcRequester{
			url:    url,
			client: client,
		},
		resolver,
		verifier,
		concurrentRequests,
		4*MB,
	)
}

// This class provides the implementation of BlockSourceRequester for BlockSourceBase
// this simplifies creating new BlockSources that satisfy the requirements down to
// writing a request function
type GrpcRequester struct {
	client pb.VimonadeServiceClient
	url    string
}

func (r *GrpcRequester) DoRequest(startOffset int64, endOffset int64) (data []byte, err error) {
	// ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	// defer cancel()

	// resp, err := r.client.Sync(ctx, &pb.SyncRequest{
	// StartOffset: startOffset,
	// EndOffset:   endOffset,
	// })
	// if err != nil {
	// return nil, fmt.Errorf("Error creating request for \"%v\": %v", r.url, err)
	// }
	//
	// buf := bytes.NewBuffer(make([]byte, 0, endOffset-startOffset))
	// _, err = buf.Read(resp.GetData())
	//
	// if err != nil {
	// err = fmt.Errorf(
	// "Failed to read response body for %v (%v-%v): %v",
	// r.url,
	// startOffset, endOffset-1,
	// err,
	// )
	// }

	// data = buf.Bytes()
	//
	// if int64(len(data)) != endOffset-startOffset {
	// err = fmt.Errorf(
	// "Unexpected response length %v (%v): %v",
	// r.url,
	// endOffset-startOffset+1,
	// len(data),
	// )
	// }

	// file, err := os.Open(r.url)
	// if err != nil {
	// return err
	// }
	// defer file.Close()
	//
	// reader := bufio.NewReader(file)
	// buffer := make([]byte, 1024)
	//
	// for {
	// n, err := reader.Read(buffer)
	// if err == io.EOF {
	// break
	// }
	//
	// if err != nil {
	// return err
	// }
	//
	// req := &pb.SendFileRequest{
	// Data: &pb.SendFileRequest_ChunkData{
	// ChunkData: buffer[:n],
	// },
	// }
	//
	// err = stream.Send(req)
	// if err != nil {
	// return err
	// }
	// }

	return
}

func (r *GrpcRequester) IsFatal(err error) bool {
	return true
}
