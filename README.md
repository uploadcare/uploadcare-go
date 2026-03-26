# Golang API client for Uploadcare

![license](https://img.shields.io/badge/license-MIT-brightgreen.svg)
[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://pkg.go.dev/github.com/uploadcare/uploadcare-go/v2/ucare)
![](https://github.com/uploadcare/uploadcare-go/workflows/test/badge.svg)

Uploadcare Golang API client that handles uploads and further operations with files by wrapping Uploadcare Upload and REST APIs.

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Useful links](#useful-links)

## Requirements

Go 1.25

## Installation

Install uploadcare-go with:

```
go get -u -v github.com/uploadcare/uploadcare-go/v2/...
```

Then import it using:

```go
import (
	"github.com/uploadcare/uploadcare-go/v2/ucare"
	"github.com/uploadcare/uploadcare-go/v2/file"
	"github.com/uploadcare/uploadcare-go/v2/group"
	"github.com/uploadcare/uploadcare-go/v2/upload"
	"github.com/uploadcare/uploadcare-go/v2/conversion"
	"github.com/uploadcare/uploadcare-go/v2/projectapi"
)
```

## Configuration

### REST & Upload API client

```go
creds := ucare.APICreds{
	SecretKey: "your-project-secret-key",
	PublicKey: "your-project-public-key",
}

conf := &ucare.Config{
	SignBasedAuthentication: true,
}

client, err := ucare.NewClient(creds, conf)
if err != nil {
	log.Fatal("creating uploadcare API client: %s", err)
}
```

### Project API client

The Project API uses bearer token authentication. Tokens can be obtained
via [Uploadcare Support](mailto:help@uploadcare.com).

```go
client, err := ucare.NewBearerClient("your-bearer-token", nil)
if err != nil {
	log.Fatal("creating project API client: %s", err)
}

projectSvc := projectapi.NewService(client)
```

## Usage

For a comprehensive list of examples, check out the [API documentation](https://pkg.go.dev/github.com/uploadcare/uploadcare-go/v2/ucare).
Below are a few usage examples:

Getting a list of files:

```go
fileSvc := file.NewService(client)

listParams := file.ListParams{
	Stored:  ucare.Bool(true),
	OrderBy: ucare.String(file.OrderByUploadedAtDesc),
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
file, err := fileSvc.Info(context.Background(), fileID, nil)
if err != nil {
	// handle error
}

if file.IsImage && file.ContentInfo != nil && file.ContentInfo.Image != nil {
	h := file.ContentInfo.Image.Height
	w := file.ContentInfo.Image.Width
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

Managing projects via the Project API:

```go
client, err := ucare.NewBearerClient("your-bearer-token", nil)
if err != nil {
	// handle error
}

projectSvc := projectapi.NewService(client)

// List all projects
projects, err := projectSvc.List(context.Background(), nil)
if err != nil {
	// handle error
}

// Get project details
proj, err := projectSvc.Get(context.Background(), projects.Results[0].PubKey)
if err != nil {
	// handle error
}
fmt.Printf("project: %s (%s)\n", proj.Name, proj.PubKey)

// Get usage metrics
usage, err := projectSvc.GetUsage(context.Background(), proj.PubKey, projectapi.UsageDateRange{
	From: "2025-01-01",
	To:   "2025-01-31",
})
if err != nil {
	// handle error
}
```

## Useful links

[Golang API client documentation](https://pkg.go.dev/github.com/uploadcare/uploadcare-go/v2/ucare)  
[Uploadcare documentation](https://uploadcare.com/docs/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)  
[Upload API reference](https://uploadcare.com/api-refs/upload-api/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)  
[REST API reference](https://uploadcare.com/api-refs/rest-api/?utm_source=github&utm_medium=referral&utm_campaign=uploadcare-go)
[Changelog](https://github.com/uploadcare/uploadcare-go/blob/master/CHANGELOG.md)  
[Contributing guide](https://github.com/uploadcare/.github/blob/master/CONTRIBUTING.md)  
[Security policy](https://github.com/uploadcare/uploadcare-go/security/policy)  
[Support](https://github.com/uploadcare/.github/blob/master/SUPPORT.md)  
