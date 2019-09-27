package ucare

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAuthCreds = errors.New("Incorrect authentication credentials")
	ErrAuthForbidden    = errors.New("Simple authentication over HTTP is " +
		"forbidden. Please, use HTTPS or signed requests instead")
	ErrInvalidVersion = errors.New("Could not satisfy the request " +
		"Accept header")
)

type RespErr struct {
	Details string `json:"detail"`
}

func (e RespErr) Error() string {
	return e.Details
}

type AuthErr struct{ RespErr }

type ThrottleErr struct {
	RetryAfter int
}

func (e ThrottleErr) Error() string {
	return fmt.Sprintf(
		"Request was throttled. Expected available in %d second",
		e.RetryAfter,
	)
}
