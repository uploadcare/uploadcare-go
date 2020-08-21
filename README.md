# Golang API client for Uploadcare

![license](https://img.shields.io/badge/license-MIT-brightgreen.svg)
[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/uploadcare/uploadcare-go/ucare)
![](https://github.com/uploadcare/uploadcare-go/workflows/test/badge.svg)

Uploadcare Golang API client that handles uploads and further operations with files by wrapping Uploadcare Upload and REST APIs.

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Useful links](#useful-links)

## Requirements

go1.13

## Installation

Install uploadcare-go with:

```
go get -u -v github.com/uploadcare/uploadcare-go/...
```

Then import it using:

```go
import (
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/group"
	"github.com/uploadcare/uploadcare-go/upload"
	"github.com/uploadcare/uploadcare-go/conversion"
)
```

## Configuration 

Creating a client:

```go
creds := ucare.APICreds{
	SecretKey: "your-project-secret-key",
	PublicKey: "your-project-public-key",
}

conf := &ucare.Config{
	SignBasedAuthentication: true,
	APIVersion:              ucare.APIv06,
}

client, err := ucare.NewClient(creds, conf)
if err != nil {
	log.Fatal("creating uploadcare API client: %s", err)
}
```

## Usage

For a comprehensive list of examples, check out the [API documentation](https://godoc.org/github.com/uploadcare/uploadcare-go/ucare).
Below are a few usage examples:

Getting a list of files:

```go
fileSvc := file.NewService(client) 

listParams := file.ListParams{
	Stored:  ucare.String(true),
	OrderBy: ucare.String(file.OrderBySizeAsc),
}

fileList, err := fileSvc.List(context.Background(), listParams)
if err != nil {
	// handle error
}
		
// getting IDs of the files
ids := make([]string, 0, 100)
for fileList.Next() {
	finfo, err :=  fileList.ReadResult()
	if err != nil {
		// handle error
	}

	ids = append(ids, finfo.ID)
}
```

Acquiring file-specific info:

```go
fileID := ids[0]
file, err := fileSvc.Info(context.Background(), fileID)
if err != nil {
	// handle error
}

if file.IsImage {
	h := file.ImageInfo.Height
	w := file.ImageInfo.Width
	fmt.Printf("image size: %dx%d\n", h, w)
}
```

Uploading a file:

```go
f, err := os.Open("file.png")
if err != nil {
	// handle error
}

uploadSvc := upload.NewService(client)

params := upload.FileParams{
	Data:        f,
	Name:        f.Name(),
	ContentType: "image/png",
}
fID, err := uploadSvc.File(context.Background(), params)
if err != nil {
	// handle error
}
```

## Useful links

[Golang API client documentation](https://godoc.org/github.com/uploadcare/uploadcare-go/ucare)  
[Uploadcare documentation](https://uploadcare.com/docs/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)  
[Upload API reference](https://uploadcare.com/api-refs/upload-api/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)  
[REST API reference](https://uploadcare.com/api-refs/rest-api/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)  
[Changelog](https://github.com/uploadcare/uploadcare-go/blob/master/CHANGELOG.md)  
[Contributing guide](https://github.com/uploadcare/.github/blob/master/CONTRIBUTING.md)  
[Security policy](https://github.com/uploadcare/uploadcare-go/security/policy)  
[Support](https://github.com/uploadcare/.github/blob/master/SUPPORT.md)  
