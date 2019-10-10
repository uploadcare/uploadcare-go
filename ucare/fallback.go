package ucare

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/uploadcare/uploadcare-go/internal/config"
)

type fallbackClient struct {
	conn *http.Client
}

func (c fallbackClient) NewRequest(
	ctx context.Context,
	endpoint config.Endpoint,
	method string,
	requrl string,
	data ReqEncoder,
) (*http.Request, error) {
	req, err := http.NewRequest(method, requrl, nil)
	if err != nil {
		return nil, err
	}
	err = data.EncodeReq(req)
	if err != nil {
		return nil, err
	}
	return req.WithContext(ctx), nil
}

func (c fallbackClient) Do(req *http.Request, _ interface{}) error {
	res, err := c.conn.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(string(data))
	}
	return nil
}
