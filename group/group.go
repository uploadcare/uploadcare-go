package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
)

// Info acquires some group-specific info
func (s service) Info(
	ctx context.Context,
	id string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(infoPathFormat, id),
		nil,
		&data,
	)
	return
}

// Delete removes a group by its id.
// This only deletes the group metadata, not the files within.
func (s service) Delete(
	ctx context.Context,
	id string,
) (err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodDelete,
		fmt.Sprintf(deletePathFormat, id),
		nil,
		nil,
	)
	return
}

// Info holds group specific information
type Info struct {
	// ID is a group identifier
	ID string `json:"id"`

	// CreatedAt is a date and time when a group was created
	CreatedAt *config.Time `json:"datetime_created"`

	// StoredAt is a date and time when a group was stored
	StoredAt *config.Time `json:"datetime_stored"`

	// FileCount is a number of files in a group
	FileCount uint64 `json:"files_count"`

	// CDNLink is a public CDN URL for a group
	CDNLink string `json:"cdn_url"`
}
