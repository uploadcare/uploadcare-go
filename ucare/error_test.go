package ucare

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("api_error", func(t *testing.T) {
		t.Parallel()
		err := APIError{StatusCode: 404, Detail: "not found"}
		assert.Equal(t, "uploadcare: HTTP 404: not found", err.Error())
	})

	t.Run("auth_error", func(t *testing.T) {
		t.Parallel()
		err := AuthError{APIError{StatusCode: 401, Detail: "invalid token"}}
		assert.Equal(t, "uploadcare: authentication failed: invalid token", err.Error())
	})

	t.Run("throttle_error", func(t *testing.T) {
		t.Parallel()
		err := ThrottleError{RetryAfter: 5}
		assert.Equal(t, "uploadcare: request throttled, retry after 5 seconds", err.Error())
	})

	t.Run("throttle_error_zero_retry", func(t *testing.T) {
		t.Parallel()
		err := ThrottleError{}
		assert.Equal(t, "uploadcare: request throttled", err.Error())
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		err := ValidationError{APIError{StatusCode: 400, Detail: "bad field"}}
		assert.Equal(t, "uploadcare: validation error: bad field", err.Error())
	})

	t.Run("forbidden_error", func(t *testing.T) {
		t.Parallel()
		err := ForbiddenError{APIError{StatusCode: 403, Detail: "denied"}}
		assert.Equal(t, "uploadcare: forbidden: denied", err.Error())
	})
}

func TestErrorsAs(t *testing.T) {
	t.Parallel()

	t.Run("api_error", func(t *testing.T) {
		t.Parallel()
		var target APIError
		err := error(APIError{StatusCode: 409, Detail: "conflict"})
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, 409, target.StatusCode)
	})

	t.Run("auth_error", func(t *testing.T) {
		t.Parallel()
		var target AuthError
		err := error(AuthError{APIError{StatusCode: 401, Detail: "bad creds"}})
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, 401, target.StatusCode)
	})

	t.Run("throttle_error", func(t *testing.T) {
		t.Parallel()
		var target ThrottleError
		err := error(ThrottleError{RetryAfter: 10})
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, 10, target.RetryAfter)
	})

	t.Run("auth_does_not_match_api_error", func(t *testing.T) {
		t.Parallel()
		var target APIError
		err := error(AuthError{APIError{StatusCode: 401, Detail: "bad creds"}})
		// AuthError embeds APIError as a value, not a pointer.
		// errors.As does not unwrap embedded value types, so this does NOT match.
		assert.False(t, errors.As(err, &target))
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		var target ValidationError
		err := error(ValidationError{APIError{StatusCode: 400, Detail: "bad input"}})
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, 400, target.StatusCode)
	})

	t.Run("forbidden_error", func(t *testing.T) {
		t.Parallel()
		var target ForbiddenError
		err := error(ForbiddenError{APIError{StatusCode: 403, Detail: "no"}})
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, 403, target.StatusCode)
	})
}
