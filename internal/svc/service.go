// Package svc holds (hides) all common service logic.
package svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/uclog"
)

// Service is intended to wrap some common service operations
type Service struct {
	client ucare.Client
	log    uclog.Logger
}

// New returns new Service instance
func New(client ucare.Client, log uclog.Logger) Service {
	return Service{client, log}
}

// List returns a list of raw results, *codec.ResultBuf later must be wrapped
// with somem concrete service type.
func (s Service) List(
	ctx context.Context,
	path string,
	params ucare.ReqEncoder,
) (*codec.ResultBuf, error) {
	if params == nil {
		return nil, errors.New("nil params passed")
	}

	method := http.MethodGet
	url := config.RESTAPIEndpoint + path

	req, err := s.client.NewRequest(ctx, method, url, params)
	if err != nil {
		return nil, err
	}

	resbuf := codec.ResultBuf{
		Ctx:       ctx,
		ReqMethod: method,
		Client:    s.client,
	}
	err = s.client.Do(req, &resbuf)
	if err != nil {
		return nil, err
	}

	return &resbuf, nil
}

var errEmptyFileID = errors.New("empty file id")

// ResourceOp operates on single resource. The response data is
// written into the resourceData param.
func (s Service) ResourceOp(
	ctx context.Context,
	method string,
	pathFormat string,
	resourceID string,
	resourceData interface{},
) error {
	if resourceID == "" {
		return errEmptyFileID
	}

	path := fmt.Sprintf(pathFormat, resourceID)
	url := config.RESTAPIEndpoint + path

	s.log.Infof("requesting: %s %s", method, url)

	req, err := s.client.NewRequest(ctx, method, url, nil)
	if err != nil {
		return err
	}

	err = s.client.Do(req, &resourceData)

	s.log.Debugf("received info: %+v", resourceData)

	return nil
}
