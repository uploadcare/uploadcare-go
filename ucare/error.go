package ucare

import (
	"errors"
	"fmt"
)

// API response errors
var (
	ErrInvalidAuthCreds = errors.New("incorrect authentication credentials")
	ErrAuthForbidden = errors.New("simple authentication over HTTP is " +
		"forbidden; use HTTPS or signed requests instead")
	ErrInvalidVersion = errors.New("this feature is not supported; " +
		"try changing the API version (see " +
		"https://uploadcare.com/api-refs/rest-api/v0.7.0/ for " +
		"which methods belong to which version)")
	ErrFileTooLarge = errors.New("direct uploads only support " +
		"files smaller than 100MB")
)

type respErr struct {
	Details string `json:"detail"`
}

// Error implements error interface
func (e respErr) Error() string {
	return e.Details
}

type authErr struct{ respErr }

type throttleErr struct {
	RetryAfter int
}

func (e throttleErr) Error() string {
	if e.RetryAfter == 0 {
		return "Request was throttled."
	}
	return fmt.Sprintf(
		"Request was throttled. Expected available in %d second",
		e.RetryAfter,
	)
}

type reqValidationErr struct{ respErr }

func (e reqValidationErr) Error() string {
	return fmt.Sprintf("Request parameters validation error: %s", e.Details)
}

type reqForbiddenErr struct{ respErr }

type unexpectedStatusErr struct {
	StatusCode int
}

func (e unexpectedStatusErr) Error() string {
	return fmt.Sprintf("unexpected HTTP status: %d", e.StatusCode)
}
