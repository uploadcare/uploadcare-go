package ucare

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"api_error", APIError{StatusCode: 404, Detail: "not found"}, "uploadcare: HTTP 404: not found"},
		{"auth_error", AuthError{APIError{StatusCode: 401, Detail: "invalid token"}}, "uploadcare: authentication failed: invalid token"},
		{"throttle_error", ThrottleError{RetryAfter: 5}, "uploadcare: request throttled, retry after 5 seconds"},
		{"throttle_error_zero", ThrottleError{}, "uploadcare: request throttled"},
		{"validation_error", ValidationError{APIError{StatusCode: 400, Detail: "bad field"}}, "uploadcare: validation error: bad field"},
		{"forbidden_error", ForbiddenError{APIError{StatusCode: 403, Detail: "denied"}}, "uploadcare: forbidden: denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestErrorsAs(t *testing.T) {
	t.Parallel()

	t.Run("errors_as_own_type", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			err    error
			target any
			check  func(t *testing.T)
		}{
			{"api_error", APIError{StatusCode: 409, Detail: "conflict"}, new(APIError), nil},
			{"auth_error", AuthError{APIError{StatusCode: 401, Detail: "bad creds"}}, new(AuthError), nil},
			{"throttle_error", ThrottleError{RetryAfter: 10}, new(ThrottleError), nil},
			{"validation_error", ValidationError{APIError{StatusCode: 400, Detail: "bad input"}}, new(ValidationError), nil},
			{"forbidden_error", ForbiddenError{APIError{StatusCode: 403, Detail: "no"}}, new(ForbiddenError), nil},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				require.True(t, errors.As(tt.err, tt.target))
			})
		}
	})

	t.Run("unwraps_to_api_error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			err        error
			wantStatus int
			wantDetail string
		}{
			{"auth", AuthError{APIError{StatusCode: 401, Detail: "bad creds"}}, 401, "bad creds"},
			{"validation", ValidationError{APIError{StatusCode: 400, Detail: "bad input"}}, 400, "bad input"},
			{"forbidden", ForbiddenError{APIError{StatusCode: 403, Detail: "no"}}, 403, "no"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				var target APIError
				require.True(t, errors.As(tt.err, &target))
				assert.Equal(t, tt.wantStatus, target.StatusCode)
				assert.Equal(t, tt.wantDetail, target.Detail)
			})
		}
	})
}
