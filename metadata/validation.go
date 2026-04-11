package metadata

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MaxKeyLength   = 64
	MaxValueLength = 512
	MaxKeysNumber  = 50
)

var (
	keyPattern         *regexp.Regexp
	ErrInvalidKey      = errors.New("metadata key must match FILE_METADATA_KEY_PATTERN and FILE_METADATA_MAX_KEY_LENGTH")
	ErrInvalidFileUUID = errors.New("file UUID must be a non-empty string without slashes or dot segments")
	ErrValueTooLong    = errors.New("metadata value exceeds FILE_METADATA_MAX_VALUE_LENGTH")
	ErrTooManyKeys     = errors.New("metadata exceeds FILE_METADATA_MAX_KEYS_NUMBER")
)

func init() {
	keyPattern = regexp.MustCompile(fmt.Sprintf(`^[\w.:-]{1,%d}$`, MaxKeyLength))
}

func validateFileUUID(fileUUID string) error {
	if fileUUID == "" || fileUUID == "." || fileUUID == ".." ||
		strings.Contains(fileUUID, "/") {
		return fmt.Errorf("%w: %q", ErrInvalidFileUUID, fileUUID)
	}
	return nil
}

func validateKey(key string) error {
	if !keyPattern.MatchString(key) {
		return fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}
	return nil
}

func validateValue(value string) error {
	if utf8.RuneCountInString(value) > MaxValueLength {
		return fmt.Errorf("%w", ErrValueTooLong)
	}
	return nil
}

func WouldExceedKeyLimit(existing map[string]string, key string) bool {
	if existing == nil {
		return false
	}
	if _, ok := existing[key]; ok {
		return false
	}
	return len(existing) >= MaxKeysNumber
}
