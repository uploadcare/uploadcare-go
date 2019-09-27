package main

import (
	"context"
	"fmt"
	"log"

	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/ucare"
)

func main() {
	creds := ucare.APICreds{
		SecretKey: "857160ed1354414a144d",
		PublicKey: "4a915d7016ad96b979d5",
	}

	//ucare.EnableLog(uclog.LevelDebug)
	//file.EnableLog(uclog.LevelDebug)

	client, err := ucare.NewClient(
		creds,
		ucare.WithAuthentication(ucare.SignBasedAuth),
		ucare.WithAPIVersion(ucare.APIv05),
	)
	if err != nil {
		log.Fatal(err)
	}

	fileSvc := file.New(client)

	params := file.ListParams{
		Limit:    ucare.Int64(3),
		Stored:   ucare.Bool(true),
		Removed:  ucare.Bool(false),
		Ordering: ucare.String(file.OrderByUploadedAtAsc),
	}

	fileList, err := fileSvc.ListFiles(context.Background(), &params)
	if err != nil {
		log.Fatal(err)
	}

	for fileList.Next() {
		finfo, err := fileList.ReadResult()
		if err != nil {
			log.Print(err)
		}

		fmt.Printf("%+v\n", finfo.ID)
	}
}
