// Package svc holds (hides) all common service logic.
package svc

import (
	"context"
	"errors"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/codec"
	"github.com/uploadcare/uploadcare-go/internal/config"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/uclog"
)

// Service is intended to wrap some common service operations
type Service struct {
	endpoint config.Endpoint

	client ucare.Client
	log    uclog.Logger
}

// New returns new Service instance
func New(
	endpoint config.Endpoint,
	client ucare.Client,
	log uclog.Logger,
) Service {
	return Service{endpoint, client, log}
}

// ErrNilParams is returned when method does not allow nil params to be passed
var ErrNilParams = errors.New("nil params passed")

// List returns a list of raw results, *codec.ResultBuf later must be wrapped
// with a concrete service type.
func (s Service) List(
	ctx context.Context,
	path string,
	params ucare.ReqEncoder,
) (*codec.ResultBuf, error) {
	method := http.MethodGet
	resbuf := codec.ResultBuf{
		Ctx:       ctx,
		ReqMethod: method,
		Client:    s.client,
	}

	return &resbuf, s.ResourceOp(ctx, method, path, params, &resbuf)
}

// ResourceOp operates on single resource. The response data is
// written into the resourceData param.
func (s Service) ResourceOp(
	ctx context.Context,
	method string,
	requrl string,
	params ucare.ReqEncoder,
	resourceData interface{},
) error {
	// shouldn't be the case
	if method == "" || requrl == "" {
		return errors.New("invalid params or method passed")
	}

	s.log.Infof("requesting: %s %s", method, requrl)

	req, err := s.client.NewRequest(ctx, s.endpoint, method, requrl, params)
	if err != nil {
		return err
	}

	err = s.client.Do(req, resourceData)
	if err != nil {
		return err
	}

	s.log.Debugf("received: %+v", resourceData)
	return nil

}
