package ucare

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleThrottle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		retryAfter     string
		cfg            *RetryConfig
		tries          int
		cancelCtx      bool
		wantOK         bool
		wantRetryAfter int // checked only when wantOK=false and cancelCtx=false
	}{
		{
			name:           "nil_config",
			retryAfter:     "5",
			cfg:            nil,
			tries:          1,
			wantRetryAfter: 5,
		},
		{
			name:       "retries_exhausted",
			retryAfter: "1",
			cfg:        &RetryConfig{MaxRetries: 2, MaxWaitSeconds: 60},
			tries:      3,
		},
		{
			name:           "retry_after_exceeds_max_wait",
			retryAfter:     "10",
			cfg:            &RetryConfig{MaxRetries: 5, MaxWaitSeconds: 3},
			tries:          1,
			wantRetryAfter: 10,
		},
		{
			name:           "no_retry_after_backoff_exceeds_max_wait",
			retryAfter:     "",
			cfg:            &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 3},
			tries:          5, // expBackoff=16 > 3
			wantRetryAfter: 16,
		},
		{
			name:           "invalid_retry_after_backoff_exceeds_max_wait",
			retryAfter:     "not-a-number",
			cfg:            &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 3},
			tries:          5, // expBackoff=16 > 3
			wantRetryAfter: 16,
		},
		{
			name:       "no_retry_after_within_max_wait",
			retryAfter: "",
			cfg:        &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 60},
			tries:      1, // expBackoff=1 <= 60
			wantOK:     true,
		},
		{
			name:       "retry_after_within_max_wait",
			retryAfter: "1",
			cfg:        &RetryConfig{MaxRetries: 5, MaxWaitSeconds: 10},
			tries:      1,
			wantOK:     true,
		},
		{
			name:      "context_cancelled",
			retryAfter: "1",
			cfg:       &RetryConfig{MaxRetries: 5, MaxWaitSeconds: 60},
			tries:     1,
			cancelCtx: true,
		},
		{
			name:       "max_wait_zero_unlimited",
			retryAfter: "",
			cfg:        &RetryConfig{MaxRetries: 10, MaxWaitSeconds: 0},
			tries:      1, // expBackoff=1; no cap
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			if tt.retryAfter != "" {
				rec.Header().Set("Retry-After", tt.retryAfter)
			}
			resp := rec.Result()

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			ok, err := handleThrottle(ctx, resp, tt.cfg, tt.tries)
			assert.Equal(t, tt.wantOK, ok)

			if tt.wantOK {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			if tt.cancelCtx {
				assert.ErrorIs(t, err, context.Canceled)
				return
			}

			if tt.wantRetryAfter > 0 {
				var te ThrottleError
				require.True(t, errors.As(err, &te))
				assert.Equal(t, tt.wantRetryAfter, te.RetryAfter)
			}
		})
	}
}

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
