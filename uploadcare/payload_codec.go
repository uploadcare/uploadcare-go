package uploadcare

import (
	"net/http"
)

type RequestEncoder interface {
	EncodeRequest(*http.Request)
}
