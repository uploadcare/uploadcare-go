package ucare

import (
	"net/http"
)

// RequestEncoder exists to encode data into prepared request.
// It may encode part of the data to the query string and other
// part into the request body
type RequestEncoder interface {
	EncodeRequest(*http.Request)
}
