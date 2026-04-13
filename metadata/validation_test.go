package metadata

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func metadataMap(n int) map[string]string {
	m := make(map[string]string, n)
	for i := range n {
		m[fmt.Sprintf("k%d", i)] = "v"
	}
	return m
}

func TestValidateFileUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid arbitrary", "my-file-id", false},
		{"empty", "", true},
		{"single dot", ".", true},
		{"dotdot", "..", true},
		{"contains slash", "abc/def", true},
		{"path traversal", "../../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateFileUUID(tt.uuid)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidFileUUID)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWouldExceedKeyLimit(t *testing.T) {
	t.Parallel()

	full := metadataMap(MaxKeysNumber)
	belowCap := metadataMap(MaxKeysNumber - 1)

	tests := []struct {
		name     string
		existing map[string]string
		key      string
		want     bool
	}{
		{"nil map", nil, "new", false},
		{"update existing at cap", full, "k0", false},
		{"new key at cap", full, "new", true},
		{"new key below cap", belowCap, "new", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, WouldExceedKeyLimit(tt.existing, tt.key))
		})
	}
}
