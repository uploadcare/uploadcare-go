/*
Package ucare provides the binding for the Uploadcare API.

	import (
		"github.com/uploadcare/uploadcare-go/ucare"
		"github.com/uploadcare/uploadcare-go/file"
		"github.com/uploadcare/uploadcare-go/upload"
	)

Construct a new Uploadcare client, then use the various domain services to
access different parts of the Uploadcare API:

	creds := ucare.APICreds{
		SecretKey: "your_secret_key",
		PublicKey: "your_public_key",
	}

	conf := &ucare.Config{
		SignBasedAuthentication: true,
	}

	client, err := ucare.NewClient(creds, conf)
	if err != nil {
		// handle error
	}

NOTE: If you want to use signature based authentication, you need to enable
it in the Uploadcare dashboard first.

Getting a list of files:

	// creating a file operations service
	fileSvc := file.NewService(client)

	listParams := &file.ListParams{
		Stored:  ucare.String(true),
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

	listParams := &file.ListParams{
		OrderBy:      ucare.String(group.OrderByCreatedAtAsc),
		Limit:        ucare.String(20),
	}

	groupList, err := groupSvc.List(context.Backgroud(), listParams)
	if err != nil {
		// handle error
	}

	// getting group IDs
	groupIDs = make([]string, 0, 100)
	for groupList.Next() {
		groupList, err := groupList.ReadResult()
		if err != nil {
			// handle error
		}
		groupIDs = append(groupIDs, groupList.ID)
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

Uploading a file

	uploadSvc := upload.NewService(client)

	file, err := os.Open("somefile.png")
	if err != nil {
		// handle error
	}

	fileParams := &upload.FileParams{
		File: file,
		ToStore: ucare.String(upload.ToStoreTrue),
	}

	fileID, err := uploadSvc.UploadFile(context.Background(), fileParams)
	if err != nil {
		// handle error
	}

*/
package ucare
