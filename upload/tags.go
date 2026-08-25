package upload

import (
	"github.com/uploadcare/uploadcare-go/v2/internal/filetag"
)

const (
	MaxTagLength = filetag.MaxLength
	MaxTagCount  = filetag.MaxCount
)

var (
	ErrTagBlank             = filetag.ErrBlank
	ErrTagTooLong           = filetag.ErrTooLong
	ErrTagInvalidCharacters = filetag.ErrInvalidCharacters
	ErrTagTooMany           = filetag.ErrTooMany
)

func normalizeUploadTags(tags *[]string) error {
	normalized, err := filetag.Normalize(*tags, filetag.MaxCount)
	if err != nil {
		return err
	}
	*tags = normalized
	return nil
}
