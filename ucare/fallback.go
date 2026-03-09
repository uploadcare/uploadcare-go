package ucare

import (
	"errors"
	"io"
	"net/http"
)

func fallbackDoFunc(client *http.Client) func(*http.Request, interface{}) error {
	return func(req *http.Request, _ interface{}) error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return errors.New(string(data))
		}
		return nil
	}
}
