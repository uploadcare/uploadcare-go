/*
Package ucare provides the binding for the Uploadcare API.

	import (
		"github.com/uploadcare/uploadcare-go/v2/ucare"
		"github.com/uploadcare/uploadcare-go/v2/file"
		"github.com/uploadcare/uploadcare-go/v2/group"
		"github.com/uploadcare/uploadcare-go/v2/upload"
		"github.com/uploadcare/uploadcare-go/v2/conversion"
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

Getting a list of files:

	// creating a file operations service
	fileSvc := file.NewService(client)

	listParams := file.ListParams{
		Stored:  ucare.Bool(true),
		OrderBy: ucare.String(file.OrderByUploadedAtDesc),
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

	if file.IsImage && file.ContentInfo != nil && file.ContentInfo.Image != nil {
		h := file.ContentInfo.Image.Height
		w := file.ContentInfo.Image.Width
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

	groupSvc := group.NewService(client)

	listParams := group.ListParams{
		OrderBy: ucare.String(group.OrderByCreatedAtAsc),
		Limit:   ucare.Uint64(20),
	}

	groupList, err := groupSvc.List(context.Background(), listParams)
	if err != nil {
		// handle error
	}

	// getting group IDs
	groupIDs := make([]string, 0, 100)
	for groupList.Next() {
		ginfo, err := groupList.ReadResult()
		if err != nil {
			// handle error
		}
		groupIDs = append(groupIDs, ginfo.ID)
	}

Getting a file group by ID:

	groupID := groupIDs[0]
	group, err := groupSvc.Info(context.Background(), groupID)
	if err != nil {
		// handle error
	}

	fmt.Printf("group %s contains %d files\n", group.ID, group.FileCount)

Uploading a file

	uploadSvc := upload.NewService(client)

	file, err := os.Open("somefile.png")
	if err != nil {
		// handle error
	}

	fileParams := upload.FileParams{
		Data:        file,
		Name:        file.Name(),
		ToStore:     ucare.String(upload.ToStoreTrue),
	}

	fileID, err := uploadSvc.File(context.Background(), fileParams)
	if err != nil {
		// handle error
	}

*/
package ucare
