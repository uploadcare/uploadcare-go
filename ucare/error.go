package ucare

import (
	"errors"
	"fmt"
)

// Sentinel errors for specific conditions
var (
	ErrInvalidAuthCreds = errors.New("incorrect authentication credentials")
	ErrAuthForbidden    = errors.New("simple authentication over HTTP is " +
		"forbidden, use HTTPS or signed requests instead")
	ErrInvalidVersion = errors.New("this feature is not supported, " +
		"try to change the version (refer to " +
		"https://uploadcare.com/api-refs/rest-api/v0.7.0/ for " +
		"more information on which methods belong to which version)")
	ErrFileTooLarge = errors.New("direct uploads only support " +
		"files smaller than 100MB")
)

// APIError represents a generic API error response.
// Callers can use errors.As to check for this type as a catch-all
// for any non-specific API error.
type APIError struct {
	StatusCode int    `json:"-"`
	Detail     string `json:"detail"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("uploadcare: HTTP %d: %s", e.StatusCode, e.Detail)
}

// AuthError represents an authentication failure (HTTP 401).
type AuthError struct{ APIError }

func (e AuthError) Error() string {
	return fmt.Sprintf("uploadcare: authentication failed: %s", e.Detail)
}

// ThrottleError represents a throttled request (HTTP 429).
type ThrottleError struct {
	RetryAfter int
}

func (e ThrottleError) Error() string {
	if e.RetryAfter == 0 {
		return "uploadcare: request throttled"
	}
	return fmt.Sprintf(
		"uploadcare: request throttled, retry after %d seconds",
		e.RetryAfter,
	)
}

// ValidationError represents a request validation failure (HTTP 400).
type ValidationError struct{ APIError }

func (e ValidationError) Error() string {
	return fmt.Sprintf("uploadcare: validation error: %s", e.Detail)
}

// ForbiddenError represents a forbidden request (HTTP 403).
type ForbiddenError struct{ APIError }

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("uploadcare: forbidden: %s", e.Detail)
}
