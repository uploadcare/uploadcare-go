package ucare

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// body-read errors are joined onto the returned error by the caller, so
// implementations can focus on parsing.
type httpErrorMapper func(statusCode int, body []byte) error

func doWithRetry(
	conn *http.Client,
	retry *RetryConfig,
	req *http.Request,
	resdata interface{},
	mapErr httpErrorMapper,
) error {
	for tries := 1; ; tries++ {
		if tries > 1 && req.GetBody != nil {
			var err error
			req.Body, err = req.GetBody()
			if err != nil {
				return err
			}
		}

		log.Debugf("making %d request: %s %s", tries, req.Method, req.URL)

		resp, err := conn.Do(req)
		if err != nil {
			return err
		}

		again, err := processResponse(resp, req, resdata, retry, tries, mapErr)
		if err != nil || !again {
			return err
		}
	}
}

func processResponse(
	resp *http.Response,
	req *http.Request,
	resdata interface{},
	retry *RetryConfig,
	tries int,
	mapErr httpErrorMapper,
) (bool, error) {
	defer drainAndClose(resp.Body)

	log.Debugf("received response: %+v", resp)

	if resp.StatusCode == http.StatusTooManyRequests {
		return handleThrottle(req.Context(), resp, retry, tries)
	}

	if resp.StatusCode >= 400 {
		body, readErr := io.ReadAll(resp.Body)
		mapped := mapErr(resp.StatusCode, body)
		if readErr != nil {
			log.Debugf("reading error response body for HTTP %d: %v", resp.StatusCode, readErr)
			return false, errors.Join(mapped, fmt.Errorf("read response body: %w", readErr))
		}
		return false, mapped
	}

	if isNilResponseData(resdata) {
		return false, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(resdata); err != nil {
		return false, err
	}
	return false, nil
}

// drainAndClose: without draining, partially-read responses force net/http
// to drop the underlying TCP connection instead of reusing it.
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
