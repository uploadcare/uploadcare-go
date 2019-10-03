package group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/config"
)

// Info acquires some group-specific info
func (s service) Info(
	ctx context.Context,
	groupID string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		fmt.Sprintf(infoPathFormat, groupID),
		nil,
		&data,
	)
	return
}

// Store marks all files in group as stored
func (s service) Store(
	ctx context.Context,
	groupID string,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		fmt.Sprintf(storePathFormat, groupID),
		nil,
		&data,
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
