package filetag

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []string
		max     int
		want    []string
		wantErr error
	}{
		{
			name:  "normalizes_and_deduplicates",
			input: []string{" Cat ", "dog", "CAT", "release.2026"},
			max:   MaxCount,
			want:  []string{"cat", "dog", "release.2026"},
		},
		{
			name:    "blank",
			input:   []string{"cat", "  "},
			max:     MaxCount,
			wantErr: ErrBlank,
		},
		{
			name:    "blank validated before other values",
			input:   []string{"invalid tag", ""},
			max:     MaxCount,
			wantErr: ErrBlank,
		},
		{
			name:    "too_long",
			input:   []string{strings.Repeat("a", MaxLength+1)},
			max:     MaxCount,
			wantErr: ErrTooLong,
		},
		{
			name:    "invalid_characters",
			input:   []string{"has space"},
			max:     MaxCount,
			wantErr: ErrInvalidCharacters,
		},
		{
			name:    "too_many_after_deduplication",
			input:   makeTags(MaxCount + 1),
			max:     MaxCount,
			wantErr: ErrTooMany,
		},
		{
			name:  "duplicates_do_not_exceed_count",
			input: append(makeTags(MaxCount), "TAG0"),
			max:   MaxCount,
			want:  makeTags(MaxCount),
		},
		{
			name:    "count reported before invalid characters",
			input:   append(makeTags(MaxCount+1), "bad tag"),
			max:     MaxCount,
			wantErr: ErrTooMany,
		},
		{
			name:    "count reported before too long",
			input:   append(makeTags(MaxCount+1), strings.Repeat("a", MaxLength+1)),
			max:     MaxCount,
			wantErr: ErrTooMany,
		},
		{
			name:  "count_limit_disabled",
			input: makeTags(MaxCount + 1),
			want:  makeTags(MaxCount + 1),
		},
		{
			name:  "nil_returns_non_nil_empty",
			input: nil,
			max:   MaxCount,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := append([]string(nil), tt.input...)
			got, err := Normalize(tt.input, tt.max)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.NotNil(t, got)
			}
			assert.Equal(t, before, tt.input)
		})
	}
}

func makeTags(count int) []string {
	tags := make([]string, count)
	for i := range count {
		tags[i] = fmt.Sprintf("tag%d", i)
	}
	return tags
}
