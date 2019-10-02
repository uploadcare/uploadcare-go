// Package group holds all primitives and logic related file entity.
//
// Individual files on Uploadcare can be joined into groups. Those can be used
// to better organize your workflow. Technically, groups are ordered lists of
// files and can hold files together with Image Transformations in their URLs.
// The most common case with creating groups is when users upload multiple files at once.
//
// NOTE: a group itself and files within that group MUST belong to the same project.
// Groups are immutable and the only way to add/remove a file is creating a new group.
//
// Groups are identified in a way similar to individual files.
// A group ID consists of a UUID followed by a “~” tilde character and a group size:
// integer number of files in group.
// For example, here is an identifier for a group holding 12 files:
//	badfc9f7-f88f-4921-9cc0-22e2c08aa2da~12
package group

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all group related API
type Service interface {
	List(context.Context, *ListParams) (*List, error)
	Info(ctx context.Context, id string) (Info, error)
	Store(ctx context.Context, id string) (Info, error)
}

type service struct {
	svc svc.Service
}

const (
	listPathFormat  = "/groups/"
	infoPathFormat  = "/groups/%s/"
	storePathFormat = "/groups/%s/storage/"
)

// OrderBy predefined constants to be used in request params
const (
	OrderByCreatedAtAsc  = "datetime_created"
	OrderByCreatedAtDesc = "-datetime_created"
)

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(client, log)}
}
