package ucare

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
		t.Run(fmt.Sprintf("attempt_%d", c.attempt), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, expBackoff(c.attempt))
		})
	}
}
