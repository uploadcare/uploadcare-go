# uploadcare-go

Go library for accessing Uploadcard API https://uploadcare.com/

```
You are viewing the source code of the upcoming v.0.1.0.
This release is still in progress.
```

### Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Usage examples](#usage-examples)

### Requirements

go1.13

### Installation

To install the library simply run:

```
go get -u -v github.com/uploadcare/uploadcare-go/...
```

### Usage examples

Getting a paginated list of files:

```go
import(
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/file"
	...
)

func main() {
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
	fileSvc := file.New(client) 

	listParams := &file.ListFilesParams{
		Stored:   ucare.String(true),
		Ordering: ucare.String(file.OrderBySizeAsc),
	}
	
	fileList, err := fileSvc.ListFiles(context.Background(), listParams)
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
	
	...
}
```

Acquiring file-specific info:

```go
	fileID := ids[0]
	file, err := fileSvc.FileInfo(context.Background(), fileID)
	if err != nil {
		// handle error
	}

	if file.IsImage {
		h := file.ImageInfo.Height
		w := file.ImageInfo.Width
		fmt.Printf("image size: %dx%d\n", h, w)
	}
```

Deleting a file:

```go
	_, err := fileSvc.DeleteFile(context.Background(), fileID)
	if err != nil {
		// handle error
	}
```

Storing a single file by ID:

```go
	_, err := fileSvc.StoreFile(context.Background(), fileID)
	if err != nil {
		// handle error
	}
```

Getting a list of groups:

```go
	groupSvc := group.New(client)
	
	groupList, err := groupSvc.ListGroups()
	if err != nil {
		// handle error
	}

	// getting group IDs
	groupIDs = make([]string, 0, groupList.Total)
	for groupList.Next() {
		glist, err := groupList.ReadResult()
		if err != nil {
			// handle error
		}
		groupIDs = append(groupIDs, glist.ID)
	}

```

Getting a file group by ID:

```go
	groupID := groupIDs[0]
	group, err := groupSvc.GroupInfo(context.Background(), groupID)
	if err != nil {
		// handle error
	}

	fmt.Printf("group %s contains %d files\n", group.ID, group.FileCount)

```

Marking all files in a group as stored:

```go
	_, err := groupSvc.StoreGroup(context.Background(), groupID)
	if err != nil {
		// handle error
	}
```
