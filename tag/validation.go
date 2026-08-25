package tag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/uploadcare/uploadcare-go/v2/internal/filetag"
)

const (
	MaxLength = filetag.MaxLength
	MaxCount  = filetag.MaxCount
)

var (
	ErrBlank             = filetag.ErrBlank
	ErrTooLong           = filetag.ErrTooLong
	ErrInvalidCharacters = filetag.ErrInvalidCharacters
	ErrTooMany           = filetag.ErrTooMany
	ErrInvalidFileUUID   = errors.New("file UUID must be a non-empty string without slashes or dot segments")
)

func validateFileUUID(fileUUID string) error {
	if fileUUID == "" || fileUUID == "." || fileUUID == ".." ||
		strings.Contains(fileUUID, "/") {
		return fmt.Errorf("%w: %q", ErrInvalidFileUUID, fileUUID)
	}
	return nil
}
