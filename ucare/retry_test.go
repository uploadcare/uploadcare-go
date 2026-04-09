package ucare

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestExpBackoff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		attempt int
		want    int
	}{
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 8},
		{5, 16},
		{6, 30}, // capped
		{7, 30},
		{10, 30},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, expBackoff(c.attempt), "attempt %d", c.attempt)
	}
}
