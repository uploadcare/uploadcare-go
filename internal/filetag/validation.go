package filetag

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MaxLength = 100
	MaxCount  = 50
)

var (
	ErrBlank             = errors.New("tag may not be blank")
	ErrTooLong           = errors.New("tag exceeds maximum length")
	ErrInvalidCharacters = errors.New("tag contains invalid characters")
	ErrTooMany           = errors.New("too many tags")
)

var validPattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

// Normalize returns a normalized, deduplicated copy of tags. A maxCount of
// zero disables count validation; the returned slice is always non-nil.
func Normalize(tags []string, maxCount int) ([]string, error) {
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			return nil, ErrBlank
		}
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))

	for _, raw := range tags {
		value := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := seen[value]; ok {
			continue
		}
		if length := utf8.RuneCountInString(value); length > MaxLength {
			return nil, fmt.Errorf(
				"%w: %d characters (maximum %d)",
				ErrTooLong,
				length,
				MaxLength,
			)
		}
		if !validPattern.MatchString(value) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidCharacters, value)
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	if maxCount > 0 && len(normalized) > maxCount {
		return nil, fmt.Errorf(
			"%w: %d (maximum %d)",
			ErrTooMany,
			len(normalized),
			maxCount,
		)
	}

	return normalized, nil
}
