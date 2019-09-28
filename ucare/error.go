package ucare

import (
	"errors"
	"fmt"
)

// API response errors
var (
	ErrInvalidAuthCreds = errors.New("Incorrect authentication credentials")
	ErrAuthForbidden    = errors.New("Simple authentication over HTTP is " +
		"forbidden. Please, use HTTPS or signed requests instead")
	ErrInvalidVersion = errors.New("Could not satisfy the request " +
		"Accept header")
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
	return fmt.Sprintf(
		"Request was throttled. Expected available in %d second",
		e.RetryAfter,
	)
}
