package ucare

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/uploadcare/uploadcare-go/internal/config"
)

type uploadClient struct {
	authFunc UploadAPIAuthFunc

	conn *http.Client
}

func newUploadClient(creds APICreds, conf *Config) Client {
	c := uploadClient{
		authFunc: simpleUploadAPIAuthFunc(creds),
		conn:     conf.HTTPClient,
	}

	if conf.SignBasedAuthentication {
		c.authFunc = signBasedUploadAPIAuthFunc(creds)
	}

	return &c
}

func (c *uploadClient) NewRequest(
	ctx context.Context,
	endpoint config.Endpoint,
	method string,
	requrl string,
	data ReqEncoder,
) (*http.Request, error) {
	requrl, err := resolveReqURL(endpoint, requrl)
	if err != nil {
		return nil, fmt.Errorf("resolving req url: %w", err)
	}
	req, err := http.NewRequest(method, requrl, nil)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, config.CtxAuthFuncKey, c.authFunc)
	req = req.WithContext(ctx)

	if data != nil {
		if err := data.EncodeReq(req); err != nil {
			return nil, err
		}
	}

	log.Debugf(
		"created new request: %s %+v %+v",
		req.Method,
		req.URL,
		req.Header,
	)

	return req, nil
}

func (c *uploadClient) Do(
	req *http.Request,
	resdata interface{},
) error {
	tries := 0
try:
	tries++

	log.Debugf("making %d request: %s %+v", tries, req.Method, req.URL)

	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}
	req.Body.Close()

	log.Debugf("received response: %+v", resp)

	switch resp.StatusCode {
	case 400, 403:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		switch resp.StatusCode {
		case 400:
			return reqValidationErr{respErr{string(data)}}
		case 403:
			return reqForbiddenErr{respErr{string(data)}}
		}
	case 413:
		return ErrFileTooLarge
	case 429:
		if tries > config.MaxThrottleRetries {
			return throttleErr{}
		}
		// retry after is not returned from the upload API
		time.Sleep(5 * time.Second)
		goto try
	}

	err = json.NewDecoder(resp.Body).Decode(&resdata)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
