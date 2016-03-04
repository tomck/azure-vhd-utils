package main

import (
	"github.com/codegangsta/cli"
	"strings"
	"log"
	"strconv"
	"runtime"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Microsoft/azure-vhd-utils-for-go/vhdcore/validator"
	"github.com/Microsoft/azure-vhd-utils-for-go/vhdcore/diskstream"
	"github.com/Microsoft/azure-vhd-utils-for-go/vhdcore/common"
	"github.com/Microsoft/azure-vhd-utils-for-go/upload"
)

func vhdUploadCmdHandler() cli.Command {
	return cli.Command{
		Name:  "upload",
		Usage:  "Upload a local VHD to Azure storage as page blob",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "localvhdpath",
				Usage: "Path to source VHD in the local machine.",
			},
			cli.StringFlag{
				Name: "stgaccountname",
				Usage: "Azure storage account name.",
			},
			cli.StringFlag{
				Name: "stgaccountkey",
				Usage: "Azure storage account key.",
			},
			cli.StringFlag{
				Name: "containername",
				Usage: "Name of the container holding destination page blob. (Default: vhds)",
			},
			cli.StringFlag{
				Name: "blobname",
				Usage: "Name of the destination page blob.",
			},
			cli.BoolFlag{
				Name:  "overwrite",
				Usage: "Overwrite the blob if already exists.",
			},
		},
		Action: func (c *cli.Context) {
			const PageBlobPageSize int64 = 2 * 1024 * 1024

			localVHDPath := c.String("localvhdpath")
			if localVHDPath == "" {
				log.Fatalln("Missing required argument --localvhdpath")
			}

			stgAccountName := c.String("stgaccountname")
			if stgAccountName == "" {
				log.Fatalln("Missing required argument --stgaccountname")
			}

			stgAccountKey := c.String("stgaccountkey")
			if stgAccountKey == "" {
				log.Fatalln("Missing required argument --stgaccountkey")
			}

			containerName := c.String("containername")
			if containerName == "" {
				containerName = "vhds"
				log.Println("Using default container 'vhds'")
			}

			blobName := c.String("blobname")
			if blobName == "" {
				log.Fatalln("Missing required argument --blobname")
			}

			if !strings.HasSuffix(strings.ToLower(blobName), ".vhd") {
				blobName = blobName + ".vhd"
			}

			parallelism := int(0)
			if c.IsSet("parallelism") {
				p, err := strconv.ParseUint(c.String("parallelism"), 10, 32)
				if err != nil {
					log.Fatalln("invalid index value --parallelism: %s", err)
				}
				parallelism = int(p)
			} else {
				parallelism = 8 * runtime.NumCPU()
				log.Printf("Using default parallelism [8*NumCPU] : %d\n", parallelism)
			}

			var err error
			if err = validator.ValidateVhd(localVHDPath); err != nil {
				log.Fatal(err)
			}

			if err = validator.ValidateVhdSize(localVHDPath); err != nil {
				log.Fatal(err)
			}

			storageClient, err := storage.NewBasicClient(stgAccountName, stgAccountKey)
			if err != nil {
				log.Fatal(err)
			}
			blobServiceClient := storageClient.GetBlobService()
			if _, err = blobServiceClient.CreateContainerIfNotExists(containerName, storage.ContainerAccessTypePrivate); err != nil {
				log.Fatal(err)
			}

			diskStream, err := diskstream.CreateNewDiskStream(localVHDPath)
			if err != nil {
				panic(err)
			}
			defer diskStream.Close()

			if err = blobServiceClient.PutPageBlob(containerName, blobName, diskStream.GetSize(), nil); err != nil {
				log.Fatal(err)
			}

			var rangesToSkip = make([]*common.IndexRange, 0)
			uploadableRanges, err := upload.LocateUploadableRanges(diskStream, rangesToSkip, PageBlobPageSize)
			if err != nil {
				log.Fatal(err)
			}

			uploadableRanges, err = upload.DetectEmptyRanges(diskStream, uploadableRanges)
			if err != nil {
				log.Fatal(err)
			}

			cxt := &upload.DiskUploadContext{
				VhdStream: diskStream,
				UploadableRanges: uploadableRanges,
				BlobServiceClient: blobServiceClient,
				ContainerName: containerName,
				BlobName: blobName,
				Parallelism: parallelism,
			}

			err = upload.Upload(cxt)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}


