package upload

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Microsoft/azure-vhd-utils-for-go/upload/concurrent"
	"github.com/Microsoft/azure-vhd-utils-for-go/upload/progress"
	"github.com/Microsoft/azure-vhd-utils-for-go/vhdcore/common"
	"github.com/Microsoft/azure-vhd-utils-for-go/vhdcore/diskstream"
	"io"
	"time"
)

// DiskUploadContext type describes VHD upload context, this includes the disk stream to read from, the ranges of
// the stream to read, the destination blob and it's container, the client to communicate with Azure storage and
// the number of parallel go-routines to use for upload.
//
type DiskUploadContext struct {
	VhdStream         *diskstream.DiskStream
	UploadableRanges  []*common.IndexRange
	BlobServiceClient storage.BlobStorageClient
	ContainerName     string
	BlobName          string
	Parallelism       int
}

// Upload uploads the disk ranges described by the parameter cxt, this parameter describes the disk stream to
// read from, the ranges of the stream to read, the destination blob and it's container, the client to communicate
// with Azure storage and the number of parallel go-routines to use for upload.
//
func Upload(cxt *DiskUploadContext) error {
	// Get the channel that contains stream of disk data to upload
	dataWithRangeChan, streamReadErrChan := GetDataWithRanges(cxt.VhdStream, cxt.UploadableRanges)

	// The channel to send upload request to load-balancer
	requtestChan := make(chan *concurrent.Request, 0)

	// Prepare and start the load-balancer that load request across 'cxt.Parallelism' workers
	loadBalancer := concurrent.NewBalancer(cxt.Parallelism)
	loadBalancer.Init()
	workerErrorChan, allWorkersFinishedChan := loadBalancer.Run(requtestChan)

	// Calculate the actual size of the data to upload
	uploadSizeInBytes := int64(0)
	for _, r := range cxt.UploadableRanges {
		uploadSizeInBytes += r.Length()
	}
	// Prepare and start the upload progress tracker
	uploadProgress := progress.NewStatus(cxt.Parallelism, 0, uploadSizeInBytes, progress.NewComputestateDefaultSize())
	progressChan := uploadProgress.Run()
	// read progress status from progress tracker and print it
	go readAndPrintProgress(progressChan)

	go func() {
		for {
			fmt.Println("FailedAfterAllRetries:", <-workerErrorChan)
		}
	}()

	var err error
L:
	for {
		select {
		case dataWithRange, ok := <-dataWithRangeChan:
			if !ok {
				close(requtestChan)
				break L
			}

			req := &concurrent.Request{
				Work: func() error {
					err := cxt.BlobServiceClient.PutPage(cxt.ContainerName, cxt.BlobName, dataWithRange.Range.Start, dataWithRange.Range.End, storage.PageWriteTypeUpdate, dataWithRange.Data)
					if err == nil {
						uploadProgress.ReportBytesProcessedCount(dataWithRange.Range.Length())
					}
					return err
				},
				ShouldRetry: func(e error) bool {
					return true
				},
				ID: dataWithRange.Range.String(),
			}

			requtestChan <- req // Send to load balancer
		case err = <-streamReadErrChan:
			close(requtestChan)
			loadBalancer.TearDownWorkers()
			break L
		}
	}

	fmt.Println("\n  Waiting for all worker to finish")
	<-allWorkersFinishedChan
	fmt.Println("\n  All workers finished")
	uploadProgress.Close()
	return err
}

// GetDataWithRanges with start reading and streaming the ranges from the disk identified by the parameter ranges.
// It returns two channels, a data channel to stream the disk ranges and a channel to send any error while reading
// the disk. On successful completion the data channel will be closed. the caller must not expect any more value in
// the data channel if the error channel is signaled.
//
func GetDataWithRanges(stream *diskstream.DiskStream, ranges []*common.IndexRange) (<-chan *DataWithRange, <-chan error) {
	dataWithRangeChan := make(chan *DataWithRange, 0)
	errorChan := make(chan error, 0)
	go func() {
		for _, r := range ranges {
			dataWithRange := &DataWithRange{
				Range: r,
				Data:  make([]byte, r.Length()),
			}
			_, err := stream.Seek(r.Start, 0)
			if err != nil {
				errorChan <- err
				return
			}
			_, err = io.ReadFull(stream, dataWithRange.Data)
			if err != nil {
				errorChan <- err
				return
			}
			dataWithRangeChan <- dataWithRange
		}
		close(dataWithRangeChan)
	}()
	return dataWithRangeChan, errorChan
}

// readAndPrintProgress reads the progress records from the given progress channel and output it. It reads the
// progress record until the channel is closed.
//
func readAndPrintProgress(progressChan <-chan *progress.Record) {
	s := time.Time{}
	fmt.Println("\nUploading the VHD..")
	for progressRecord := range progressChan {
		t := s.Add(progressRecord.RemainingDuration)
		fmt.Printf("\r Completed: %3d%% RemainingTime: %02dh:%02dm:%02ds Throughput: %d MB/sec",
			int(progressRecord.PercentComplete),
			t.Hour(), t.Minute(), t.Second(),
			int(progressRecord.AverageThroughputMBPerSecond),
		)
	}
}
