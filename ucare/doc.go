/*
Package ucare provides the binding for the Uploadcare API.

	import (
		"github.com/uploadcare/uploadcare-go/ucare"
		"github.com/uploadcare/uploadcare-go/file"
	)

Construct a new Uploadcare client, then use the various domain services to
access different parts of the Uploadcare API:

	creds := ucare.APICreds{
		SecretKey: "your_secret_key",
		PublicKey: "your_public_key",
	}

	client, err := ucare.NewClient(creds)
	if err != nil {
		// handle error
	}

To authenticate your account, every request made MUST be signed.
There are two available auth functions for that:

	ucare.SimpleAuth (default)
	ucare.SignBasedAuth

NOTE: If you want to use SignBasedAuth you need to enable it in the Uploadcare
dashboard first.

Getting a paginated list of files:

	// creating a file operations service
	fileSvc := file.NewService(client)

	listParams := &file.ListParams{
		Stored:   ucare.String(true),
		Ordering: ucare.String(file.OrderBySizeAsc),
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

Acquiring file-specific info:

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

Deleting a file

	_, err := fileSvc.Delete(context.Background(), fileID)
	if err != nil {
		// handle error
	}

Storing a single file by ID:

	_, err := fileSvc.Store(context.Background(), fileID)
	if err != nil {
		// handle error
	}

Getting a list of groups:

	groupSvc := group.New(client)

	groupList, err := groupSvc.List()
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

Getting a file group by ID:

	groupID := groupIDs[0]
	group, err := groupSvc.Info(context.Background(), groupID)
	if err != nil {
		// handle error
	}

	fmt.Printf("group %s contains %d files\n", group.ID, group.FileCount)

Marking all files in a group as stored:

	_, err := groupSvc.Store(context.Background(), groupID)
	if err != nil {
		// handle error
	}

*/
package ucare
