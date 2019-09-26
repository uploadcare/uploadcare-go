package uploadcare

import (
	"io"
	"net/http"
)

type RequestEncoder interface {
	EncodeRequest(*http.Request)
}

type RespBodyDecoder interface {
	DecodeRespBody(io.Reader) error
}
