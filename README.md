# uploadcare-go

<img 
	align="right"
	width="64"
	height="64"
	src="https://ucarecdn.com/2f4864b7-ed0e-4411-965b-8148623aa680/uploadcare-logo-mark.svg"
	alt=""
/>

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/uploadcare/uploadcare-go/ucare)
![](https://github.com/uploadcare/uploadcare-go/workflows/test/badge.svg)

Go library for accessing Uploadcard API https://uploadcare.com/

### Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Documentation](#documentation)

### Requirements

go1.13

### Installation

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

### Documentation

For a comprehensive list of examples, check out the [API documentation](https://godoc.org/github.com/uploadcare/uploadcare-go/ucare).
Below are a few usage examples:

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

----


MIT License. Copyright (c) 2019 Uploadcare
