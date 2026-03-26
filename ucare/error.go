package ucare

import (
	"errors"
	"fmt"
)

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

type APIError struct {
	StatusCode int    `json:"-"`
	Detail     string `json:"detail"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("uploadcare: HTTP %d: %s", e.StatusCode, e.Detail)
}

type AuthError struct{ APIError }

func (e AuthError) Error() string {
	return fmt.Sprintf("uploadcare: authentication failed: %s", e.Detail)
}

func (e AuthError) Unwrap() error { return e.APIError }

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

type ValidationError struct{ APIError }

func (e ValidationError) Error() string {
	return fmt.Sprintf("uploadcare: validation error: %s", e.Detail)
}

func (e ValidationError) Unwrap() error { return e.APIError }

type ForbiddenError struct{ APIError }

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("uploadcare: forbidden: %s", e.Detail)
}

func (e ForbiddenError) Unwrap() error { return e.APIError }

// ProjectAPIError represents an error response from the Project API.
// The Project API returns errors as {"message": "...", "code": "..."}.
type ProjectAPIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Code       string `json:"code"`
}

func (e ProjectAPIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("uploadcare: project api: HTTP %d: %s (%s)", e.StatusCode, e.Message, e.Code)
	}
	return fmt.Sprintf("uploadcare: project api: HTTP %d: %s", e.StatusCode, e.Message)
}
