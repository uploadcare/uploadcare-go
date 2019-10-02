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

// ErrNilParams is returned when method does not allow nil params to be passed
var ErrNilParams = errors.New("nil params passed")

// List returns a list of raw results, *codec.ResultBuf later must be wrapped
// with somem concrete service type.
func (s Service) List(
	ctx context.Context,
	path string,
	params ucare.ReqEncoder,
) (*codec.ResultBuf, error) {
	if params == nil {
		return nil, ErrNilParams
	}

	endpoint := config.RESTAPIEndpoint
	method := http.MethodGet

	req, err := s.client.NewRequest(ctx, endpoint, method, path, params)
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

	endpoint := config.RESTAPIEndpoint
	requrl := fmt.Sprintf(pathFormat, resourceID)

	s.log.Infof("requesting: %s %s", method, requrl)

	req, err := s.client.NewRequest(ctx, endpoint, method, requrl, nil)
	if err != nil {
		return err
	}

	err = s.client.Do(req, &resourceData)

	s.log.Debugf("received info: %+v", resourceData)

	return nil
}
