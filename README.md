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
	"github.com/uploadcare/uploadcare-go/v2/metadata"
	"github.com/uploadcare/uploadcare-go/v2/tag"
	"github.com/uploadcare/uploadcare-go/v2/addon"
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

conf, err := ucare.NewConfig(creds, ucare.WithSignBasedAuthentication())
if err != nil {
	log.Fatalf("creating uploadcare API config: %s", err)
}

client, err := ucare.NewClient(creds, conf)
if err != nil {
	log.Fatalf("creating uploadcare API client: %s", err)
}
```

`NewConfig` accepts additional options:

```go
conf, err := ucare.NewConfig(creds,
	ucare.WithSignBasedAuthentication(),
	ucare.WithUserAgent("my-app/1.0.0"),                  // appended to the default User-Agent
	ucare.WithRetry(&ucare.RetryConfig{MaxRetries: 3}),   // retry throttled (429) requests; off by default
	ucare.WithCDNBase("https://cdn.example.com"),         // override the per-project CDN domain
)
```

By default the CDN base URL is derived automatically from the public key, and
URLs returned by the API (`file.Info.OriginalFileURL`, `group.Info.CDNLink`,
`upload.GroupInfo.CDNLink`) are rewritten to point at it.

### Project API client

The Project API uses bearer token authentication. Tokens can be obtained
via [Uploadcare Support](mailto:help@uploadcare.com).

```go
conf := ucare.NewBearerConfig()
client, err := ucare.NewBearerClient("your-bearer-token", conf)
if err != nil {
	log.Fatalf("creating project API client: %s", err)
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

Searching files:

```go
searchParams := file.SearchParams{
	Query:   "invoice",
	IsImage: ucare.Bool(false),
	Sort:    []file.SearchSort{file.SortByUploadedAtDesc},
}

results, err := fileSvc.Search(context.Background(), searchParams)
if err != nil {
	// handle error
}

fmt.Printf("found %d matches\n", results.Total())
for results.Next() {
	match, err := results.ReadResult()
	if err != nil {
		// handle error
	}

	fmt.Println(match.ID, match.OriginalFileName)
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

`Upload` picks between direct and multipart uploads automatically based on file
size (multipart above 10MB by default) and accepts custom metadata:

```go
f, err := os.Open("large-video.mp4")
if err != nil {
	// handle error
}

info, err := uploadSvc.Upload(context.Background(), upload.UploadParams{
	Data:        f,
	Name:        f.Name(),
	ContentType: "video/mp4", // required for the multipart path (files > 10MB)
	Metadata:    map[string]string{"source": "import"},
	Tags:        []string{"video", "import"},
})
if err != nil {
	// handle error
}
```

Working with per-file metadata:

```go
metaSvc := metadata.NewService(client)

_, err := metaSvc.Set(context.Background(), fileID, "source", "import")
if err != nil {
	// handle error
}

all, err := metaSvc.List(context.Background(), fileID)
if err != nil {
	// handle error
}
fmt.Printf("metadata: %v\n", all)
```

Working with per-file tags:

```go
tagSvc := tag.NewService(client)

change, err := tagSvc.Update(context.Background(), fileID, tag.UpdateParams{
	Add:    []string{"featured", "Summer"},
	Delete: []string{"draft"},
})
if err != nil {
	// handle error
}
fmt.Printf("tags: %v\n", change.Tags)

// Replace the complete tag list. Passing nil or an empty slice clears it.
_, err = tagSvc.Replace(context.Background(), fileID, []string{"approved"})
```

Executing an add-on (e.g. background removal) and polling for the result:

```go
addonSvc := addon.NewService(client)

exec, err := addonSvc.Execute(context.Background(), addon.AddonRemoveBG, addon.ExecuteParams{
	Target: fileID,
})
if err != nil {
	// handle error
}

status, err := addonSvc.Status(context.Background(), addon.AddonRemoveBG, exec.RequestID)
if err != nil {
	// handle error
}
fmt.Printf("addon status: %s\n", status.Status)
```

Managing projects via the Project API:

```go
conf := ucare.NewBearerConfig()
client, err := ucare.NewBearerClient("your-bearer-token", conf)
if err != nil {
	log.Fatalf("creating project API client: %s", err)
}

projectSvc := projectapi.NewService(client)

projects, err := projectSvc.List(context.Background(), nil)
if err != nil {
	log.Fatalf("listing projects: %s", err)
}
if !projects.Next() {
	log.Fatal("no projects found")
}
firstProject, err := projects.ReadResult()
if err != nil {
	log.Fatalf("reading first project: %s", err)
}

proj, err := projectSvc.Get(context.Background(), firstProject.PubKey)
if err != nil {
	log.Fatalf("getting project: %s", err)
}
fmt.Printf("project: %s (%s)\n", proj.Name, proj.PubKey)

usage, err := projectSvc.GetUsage(context.Background(), proj.PubKey, projectapi.UsageDateRange{
	From: "2025-01-01",
	To:   "2025-01-31",
})
if err != nil {
	log.Fatalf("getting usage: %s", err)
}
fmt.Printf("usage days: %d\n", len(usage.Data))
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
