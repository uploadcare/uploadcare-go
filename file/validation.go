package file

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MinSearchQueryLength = 4
	DefaultSearchLimit   = 20
	MaxSearchLimit       = 100
	MaxSearchOffsetLimit = 1000
	MaxSearchSortKeys    = 4
)

var (
	ErrSearchNoCondition         = errors.New("search requires at least one of query, phrase, exact, datetime_uploaded, size, is_image or tags")
	ErrSearchQueryTooShort       = errors.New("search query must be at least 4 characters")
	ErrSearchPhraseTooShort      = errors.New("search phrase field must be at least 4 characters")
	ErrSearchEmptyPhrase         = errors.New("search phrase requires at least one field")
	ErrSearchPhraseExactConflict = errors.New("search field cannot appear in both phrase and exact")
	ErrSearchInvalidInclude      = errors.New("search include supports only appdata")
	ErrSearchInvalidExactKey     = errors.New("search exact key must be uuid, detected_mime_type, original_filename or metadata[<key>]")
	ErrSearchEmptyExactValue     = errors.New("search exact values must not be empty")
	ErrSearchEmptyDatetime       = errors.New("search datetime_uploaded requires at least one operator")
	ErrSearchEmptySize           = errors.New("search size requires at least one operator")
	ErrSearchEmptyTags           = errors.New("search tags requires at least one tag")
	ErrSearchBlankTagValue       = errors.New("search tag value must not be blank")
	ErrSearchLimitOutOfRange     = errors.New("search limit must be between 1 and 100")
	ErrSearchOffsetTooLarge      = errors.New("search offset plus limit must not exceed 1000")
	ErrSearchTooManySortKeys     = errors.New("search sort accepts at most 4 keys")
	ErrSearchInvalidSortKey      = errors.New("search sort key is not supported")
	ErrSearchDuplicateSortKey    = errors.New("search sort keys must be unique and cannot include both directions of the same key")
)

var (
	searchMetadataExactKeyPattern = regexp.MustCompile(`^metadata\[[\w.:-]{1,64}\]$`)
	searchSortKeys                = map[SearchSort]struct{}{
		SortByScore:                {},
		SortByScoreDesc:            {},
		SortByUploadedAt:           {},
		SortByUploadedAtDesc:       {},
		SortBySize:                 {},
		SortBySizeDesc:             {},
		SortByOriginalFilename:     {},
		SortByOriginalFilenameDesc: {},
	}
)

func validateSearchParams(p SearchParams) error {
	if err := validateSearchFilters(p); err != nil {
		return err
	}

	if !p.hasCondition() {
		return ErrSearchNoCondition
	}

	if p.Include != nil && *p.Include != SearchIncludeAppData {
		return fmt.Errorf("%w: %q", ErrSearchInvalidInclude, *p.Include)
	}

	if p.Query != "" && utf8.RuneCountInString(p.Query) < MinSearchQueryLength {
		return fmt.Errorf("%w: %q", ErrSearchQueryTooShort, p.Query)
	}

	if p.Phrase != nil {
		fields := []string{
			p.Phrase.OriginalFilename,
			p.Phrase.Metadata,
			p.Phrase.DetectedMimeType,
		}
		for _, v := range fields {
			if v != "" && utf8.RuneCountInString(v) < MinSearchQueryLength {
				return fmt.Errorf("%w: %q", ErrSearchPhraseTooShort, v)
			}
		}
		if p.Phrase.OriginalFilename != "" && len(p.Exact[SearchExactKeyOriginalFilename]) > 0 {
			return fmt.Errorf("%w: %q", ErrSearchPhraseExactConflict, SearchExactKeyOriginalFilename)
		}
		if p.Phrase.DetectedMimeType != "" && len(p.Exact[SearchExactKeyDetectedMimeType]) > 0 {
			return fmt.Errorf("%w: %q", ErrSearchPhraseExactConflict, SearchExactKeyDetectedMimeType)
		}
		if p.Phrase.Metadata != "" {
			for key := range p.Exact {
				if strings.HasPrefix(key, "metadata[") {
					return fmt.Errorf("%w: %q", ErrSearchPhraseExactConflict, key)
				}
			}
		}
	}

	if p.Limit != nil && (*p.Limit < 1 || *p.Limit > MaxSearchLimit) {
		return fmt.Errorf("%w: %d", ErrSearchLimitOutOfRange, *p.Limit)
	}

	limit := uint64(DefaultSearchLimit)
	if p.Limit != nil {
		limit = *p.Limit
	}
	var offset uint64
	if p.Offset != nil {
		offset = *p.Offset
	}
	// Compare without adding: offset is unbounded and offset+limit could
	// overflow uint64 and wrap past the check. limit is <= MaxSearchLimit here,
	// so MaxSearchOffsetLimit-limit does not underflow.
	if offset > MaxSearchOffsetLimit-limit {
		return fmt.Errorf("%w: %d", ErrSearchOffsetTooLarge, offset)
	}

	return validateSearchSort(p.Sort)
}

func validateSearchSort(sort []SearchSort) error {
	if len(sort) > MaxSearchSortKeys {
		return fmt.Errorf("%w: %d", ErrSearchTooManySortKeys, len(sort))
	}
	seen := make(map[SearchSort]struct{}, len(sort))
	for _, key := range sort {
		if _, ok := searchSortKeys[key]; !ok {
			return fmt.Errorf("%w: %q", ErrSearchInvalidSortKey, key)
		}
		base := SearchSort(strings.TrimPrefix(string(key), "-"))
		if _, ok := seen[base]; ok {
			return fmt.Errorf("%w: %q", ErrSearchDuplicateSortKey, key)
		}
		seen[base] = struct{}{}
	}
	return nil
}

func validateSearchFilters(p SearchParams) error {
	if p.Phrase != nil && !p.Phrase.hasValue() {
		return ErrSearchEmptyPhrase
	}
	if p.DatetimeUploaded != nil && !p.DatetimeUploaded.hasValue() {
		return ErrSearchEmptyDatetime
	}
	if p.Size != nil && !p.Size.hasValue() {
		return ErrSearchEmptySize
	}
	if p.Tags != nil && !p.Tags.hasValue() {
		return ErrSearchEmptyTags
	}
	if p.Tags != nil {
		if err := validateSearchTags(p.Tags); err != nil {
			return err
		}
	}
	return validateSearchExact(p.Exact)
}

func validateSearchTags(tags *SearchTags) error {
	all := append(append(tags.Any, tags.All...), tags.None...)
	for _, tag := range all {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("%w: %q", ErrSearchBlankTagValue, tag)
		}
	}
	return nil
}

func validateSearchExact(exact map[string][]string) error {
	for key, values := range exact {
		if !isValidSearchExactKey(key) {
			return fmt.Errorf("%w: %q", ErrSearchInvalidExactKey, key)
		}
		if len(values) == 0 {
			return fmt.Errorf("%w: %q", ErrSearchEmptyExactValue, key)
		}
		for _, value := range values {
			if value == "" {
				return fmt.Errorf("%w: %q", ErrSearchEmptyExactValue, key)
			}
		}
	}
	return nil
}

func isValidSearchExactKey(key string) bool {
	switch key {
	case SearchExactKeyUUID, SearchExactKeyDetectedMimeType, SearchExactKeyOriginalFilename:
		return true
	default:
		return searchMetadataExactKeyPattern.MatchString(key)
	}
}

func (p SearchParams) hasCondition() bool {
	return p.Query != "" ||
		p.Phrase.hasValue() ||
		hasExactCondition(p.Exact) ||
		p.DatetimeUploaded.hasValue() ||
		p.Size.hasValue() ||
		p.IsImage != nil ||
		p.Tags.hasValue()
}

func (p *SearchPhrase) hasValue() bool {
	return p != nil &&
		(p.OriginalFilename != "" || p.Metadata != "" || p.DetectedMimeType != "")
}

func (d *SearchDatetime) hasValue() bool {
	return d != nil && (d.Gt != nil || d.Gte != nil || d.Lt != nil || d.Lte != nil)
}

func (s *SearchSize) hasValue() bool {
	return s != nil && (s.Gt != nil || s.Gte != nil || s.Lt != nil || s.Lte != nil)
}

func (t *SearchTags) hasValue() bool {
	return t != nil && (len(t.Any) > 0 || len(t.All) > 0 || len(t.None) > 0)
}

func hasExactCondition(exact map[string][]string) bool {
	for _, v := range exact {
		if len(v) > 0 {
			return true
		}
	}
	return false
}
