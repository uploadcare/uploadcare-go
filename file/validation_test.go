package file

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

func TestValidateSearchParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  SearchParams
		wantErr error
	}{
		{
			name:    "no condition",
			params:  SearchParams{},
			wantErr: ErrSearchNoCondition,
		},
		{
			name:    "limit only is not a condition",
			params:  SearchParams{Limit: ucare.Uint64(20)},
			wantErr: ErrSearchNoCondition,
		},
		{
			name:    "empty phrase is not a condition",
			params:  SearchParams{Phrase: &SearchPhrase{}},
			wantErr: ErrSearchEmptyPhrase,
		},
		{
			name:    "empty size is not a condition",
			params:  SearchParams{Size: &SearchSize{}},
			wantErr: ErrSearchEmptySize,
		},
		{
			name:    "empty datetime is not a condition",
			params:  SearchParams{DatetimeUploaded: &SearchDatetime{}},
			wantErr: ErrSearchEmptyDatetime,
		},
		{
			name:    "empty exact map is not a condition",
			params:  SearchParams{Exact: map[string][]string{}},
			wantErr: ErrSearchNoCondition,
		},
		{
			name:    "exact with empty values is not a condition",
			params:  SearchParams{Exact: map[string][]string{SearchExactKeyUUID: {}}},
			wantErr: ErrSearchEmptyExactValue,
		},
		{
			name:   "tags any only is a condition",
			params: SearchParams{Tags: &SearchTags{Any: []string{"urgent"}}},
		},
		{
			name:   "tags all only is a condition",
			params: SearchParams{Tags: &SearchTags{All: []string{"approved"}}},
		},
		{
			name:   "tags none only is a condition",
			params: SearchParams{Tags: &SearchTags{None: []string{"archived"}}},
		},
		{
			name:    "empty tags is not a condition",
			params:  SearchParams{Tags: &SearchTags{}},
			wantErr: ErrSearchEmptyTags,
		},
		{
			name:    "blank tag in any",
			params:  SearchParams{Tags: &SearchTags{Any: []string{"cat", ""}}},
			wantErr: ErrSearchBlankTagValue,
		},
		{
			name:    "whitespace-only tag in all",
			params:  SearchParams{Tags: &SearchTags{All: []string{"  "}}},
			wantErr: ErrSearchBlankTagValue,
		},
		{
			name:    "blank tag in none",
			params:  SearchParams{Tags: &SearchTags{None: []string{"archived", ""}}},
			wantErr: ErrSearchBlankTagValue,
		},
		{
			name:   "tags with a primary condition",
			params: SearchParams{IsImage: ucare.Bool(true), Tags: &SearchTags{Any: []string{"urgent"}}},
		},
		{
			name:   "valid query",
			params: SearchParams{Query: "invoice"},
		},
		{
			name:    "query too short",
			params:  SearchParams{Query: "inv"},
			wantErr: ErrSearchQueryTooShort,
		},
		{
			name:   "valid phrase",
			params: SearchParams{Phrase: &SearchPhrase{OriginalFilename: "report"}},
		},
		{
			name:    "phrase field too short",
			params:  SearchParams{Phrase: &SearchPhrase{Metadata: "ab"}},
			wantErr: ErrSearchPhraseTooShort,
		},
		{
			name: "phrase and exact conflict on original_filename",
			params: SearchParams{
				Phrase: &SearchPhrase{OriginalFilename: "report"},
				Exact:  map[string][]string{SearchExactKeyOriginalFilename: {"report.pdf"}},
			},
			wantErr: ErrSearchPhraseExactConflict,
		},
		{
			name: "phrase and exact conflict on detected_mime_type",
			params: SearchParams{
				Phrase: &SearchPhrase{DetectedMimeType: "image"},
				Exact:  map[string][]string{SearchExactKeyDetectedMimeType: {"image/png"}},
			},
			wantErr: ErrSearchPhraseExactConflict,
		},
		{
			name: "phrase and exact on different fields",
			params: SearchParams{
				Phrase: &SearchPhrase{OriginalFilename: "invoice"},
				Exact:  map[string][]string{SearchExactKeyDetectedMimeType: {"application/pdf"}},
			},
		},
		{
			name: "phrase metadata conflicts with exact metadata key",
			params: SearchParams{
				Phrase: &SearchPhrase{Metadata: "canon"},
				Exact:  map[string][]string{"metadata[camera]": {"Canon"}},
			},
			wantErr: ErrSearchPhraseExactConflict,
		},
		{
			name: "phrase metadata with exact non-metadata key is allowed",
			params: SearchParams{
				Phrase: &SearchPhrase{Metadata: "canon"},
				Exact:  map[string][]string{SearchExactKeyUUID: {"some-uuid"}},
			},
		},
		{
			name:   "is_image false is a condition",
			params: SearchParams{IsImage: ucare.Bool(false)},
		},
		{
			name:   "size only",
			params: SearchParams{Size: &SearchSize{Gte: ucare.Uint64(1000)}},
		},
		{
			name:   "exact only",
			params: SearchParams{Exact: map[string][]string{SearchExactKeyUUID: {"x"}}},
		},
		{
			name:   "metadata exact only",
			params: SearchParams{Exact: map[string][]string{"metadata[camera]": {"Canon"}}},
		},
		{
			name:    "invalid exact key",
			params:  SearchParams{Exact: map[string][]string{"metadata[]": {"Canon"}}},
			wantErr: ErrSearchInvalidExactKey,
		},
		{
			name:    "exact empty string value",
			params:  SearchParams{Exact: map[string][]string{SearchExactKeyOriginalFilename: {""}}},
			wantErr: ErrSearchEmptyExactValue,
		},
		{
			name:    "invalid include",
			params:  SearchParams{Query: "invoice", Include: ucare.String("metadata")},
			wantErr: ErrSearchInvalidInclude,
		},
		{
			name:    "limit zero",
			params:  SearchParams{Query: "invoice", Limit: ucare.Uint64(0)},
			wantErr: ErrSearchLimitOutOfRange,
		},
		{
			name:    "limit too high",
			params:  SearchParams{Query: "invoice", Limit: ucare.Uint64(101)},
			wantErr: ErrSearchLimitOutOfRange,
		},
		{
			name:    "offset plus limit too large",
			params:  SearchParams{Query: "invoice", Limit: ucare.Uint64(100), Offset: ucare.Uint64(950)},
			wantErr: ErrSearchOffsetTooLarge,
		},
		{
			name:    "offset plus default limit too large",
			params:  SearchParams{Query: "invoice", Offset: ucare.Uint64(990)},
			wantErr: ErrSearchOffsetTooLarge,
		},
		{
			name:   "offset plus limit at boundary",
			params: SearchParams{Query: "invoice", Limit: ucare.Uint64(100), Offset: ucare.Uint64(900)},
		},
		{
			name:    "offset does not overflow past the window check",
			params:  SearchParams{Query: "invoice", Offset: ucare.Uint64(math.MaxUint64)},
			wantErr: ErrSearchOffsetTooLarge,
		},
		{
			name:   "four sort keys",
			params: SearchParams{Query: "invoice", Sort: []SearchSort{SortByScore, SortBySize, SortByUploadedAt, SortByOriginalFilename}},
		},
		{
			name:    "too many sort keys",
			params:  SearchParams{Query: "invoice", Sort: []SearchSort{SortByScore, SortBySize, SortByUploadedAt, SortByOriginalFilename, SortByScoreDesc}},
			wantErr: ErrSearchTooManySortKeys,
		},
		{
			name:    "invalid sort key",
			params:  SearchParams{Query: "invoice", Sort: []SearchSort{"created"}},
			wantErr: ErrSearchInvalidSortKey,
		},
		{
			name:    "duplicate sort keys",
			params:  SearchParams{Query: "invoice", Sort: []SearchSort{SortBySize, SortBySize}},
			wantErr: ErrSearchDuplicateSortKey,
		},
		{
			name:    "both directions of the same sort key",
			params:  SearchParams{Query: "invoice", Sort: []SearchSort{SortBySize, SortBySizeDesc}},
			wantErr: ErrSearchDuplicateSortKey,
		},
		{
			name:   "unicode query counts runes",
			params: SearchParams{Query: strings.Repeat("☃", 4)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateSearchParams(tt.params)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
