// Package project holds all primitives and logic around the project resource.
package project

import (
	"context"

	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/internal/svc"
	"github.com/uploadcare/uploadcare-go/ucare"
)

// Service describes all project related API
type Service interface {
	Info(context.Context) (Info, error)
}

type service struct {
	svc svc.Service
}

const (
	infoPathFormat = "/project/"
)

// NewService returns new instance of the Service
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
