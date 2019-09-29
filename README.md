# uploadcare-go

Go library for accessing Uploadcard API https://uploadcare.com/

```
You are viewing the source code of the upcoming v0.1.0.
This release is still in progress.
```

### Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Documentation](#documentation)

### Requirements

go1.13

### Installation

To install the library simply run:

```
go get -u -v github.com/uploadcare/uploadcare-go/...
```

### Documentation

For details on all the functionality in this library, see the
[GoDoc](https://godoc.org/github.com/uploadcare/uploadcare-go/ucare)
documentation. Below are a few usage examples.

```go
import(
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/file"
)
```

Getting a paginated list of files:

```go
	// your project credentials
	creds := ucare.APICreds{
		SecretKey: "your-project-secret-key",
		PublicKey: "your-project-public-key",
	}

	// creating underlying client for API calls and authentication
	client, err := ucare.NewClient(
		creds,
		WithAuthentication(ucare.WithSignBasedAuth),
	)
	if err != nil {
		log.Fatal("creating uploadcare API client: %s", err)
	}

	// creating a file operations service
	fileSvc := file.NewService(client) 

	listParams := &file.ListParams{
		Stored:   ucare.String(true),
		OrderBy: ucare.String(file.OrderBySizeAsc),
	}
	
	fileList, err := fileSvc.List(context.Background(), listParams)
	if err != nil {
		// handle error
	}
			
	// getting IDs for the first 100 files
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
