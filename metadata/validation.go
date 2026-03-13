package metadata

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	keyPattern = regexp.MustCompile(`^[-_.:A-Za-z0-9]{1,64}$`)
	// ErrInvalidKey reports metadata keys that do not match the documented format.
	ErrInvalidKey = errors.New("metadata key must match ^[-_.:A-Za-z0-9]{1,64}$")
)

func validateKey(key string) error {
	if !keyPattern.MatchString(key) {
		return fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}
	return nil
}
