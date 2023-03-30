package utils

import (
	"context"
	"io"

	"github.com/99designs/gqlgen/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

//this package would contain a buffered channel which would upload video content onto cloud

type UploadRequest struct {
	File   *graphql.Upload
	Bucket string
}

var UploaderChan = make(chan UploadRequest, 10)
var ErrorChan = make(chan error)

func init() {
	go func() {
		ctx := context.Background()
		for {
			req := <-UploaderChan
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Got error while uploading videos to vendor: %v", err)
				ErrorChan <- err
			}

			writer, err := storageC.UploadToGCS(ctx, req.Bucket)
			if err != nil {
				log.Errorf("Got error while uploading videos to vendor: %v", err)
				ErrorChan <- err
			}

			// read the file in chunks and upload incrementally
			// create chunks of 10mb
			buf := make([]byte, 10*1024*1024)
			for {
				n, err := req.File.File.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Errorf("Failed to read file: %v", err.Error())
					break
				}

				_, err = writer.Write(buf[:n])
				if err != nil {
					log.Errorf("Failed to upload file: %v", err.Error())
					break
				}
			}
			err = writer.Close()
			if err != nil {
				log.Errorf("Failed to close writer: %v", err.Error())
				ErrorChan <- err
			}
		}
	}()
}

func SenUploadRequestToQueue(ctx context.Context, file *graphql.Upload, bucketPath string) error {
	//create message to be sent to channel
	uploadRequest := UploadRequest{
		File:   file,
		Bucket: bucketPath,
	}

	//send to channel
	UploaderChan <- uploadRequest
	return nil
}
