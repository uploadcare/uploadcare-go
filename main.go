package main

import (
	"context"
	"fmt"
	"log"

	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/uclog"
)

func main() {
	creds := ucare.APICreds{
		SecretKey: "857160ed1354414a144d",
		PublicKey: "4a915d7016ad96b979d5",
	}

	ucare.EnableLog(uclog.LevelDebug)
	file.EnableLog(uclog.LevelDebug)

	client, err := ucare.NewClient(
		creds,
		ucare.WithAuthentication(ucare.SignBasedAuth),
	)
	if err != nil {
		log.Fatal(err)
	}

	fileSvc := file.New(client)

	params := file.ListParams{
		Limit:    ucare.Int64(1000),
		Stored:   ucare.Bool(true),
		Removed:  ucare.Bool(false),
		Ordering: ucare.String(file.OrderByUploadedAtAsc),
	}

	fileList, err := fileSvc.ListFiles(context.Background(), &params)
	if err != nil {
		log.Fatal(err)
	}

	fileIDs := make([]string, 0, 100)
	for fileList.Next() {
		finfo, err := fileList.ReadResult()
		if err != nil {
			log.Print(err)
		}

		fmt.Printf("%+v\n", finfo)
		fileIDs = append(fileIDs, finfo.ID)

		if len(fileIDs) == 100 {
			break
		}
	}

	fmt.Println(fileIDs)
}
