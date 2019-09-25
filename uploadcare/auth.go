package uploadcare

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type authFunc func(APICreds, *http.Request)

const (
	simpleAuthScheme    = "Uploadcare.Simple"
	signBasedAuthScheme = "Uploadcare"

	dateHeaderFormat = time.RFC1123
)

var (
	authHeaderKey = http.CanonicalHeaderKey("Authorization")

	dateHeaderLocation = time.FixedZone("GMT", 0)
)

// SimpleAuth provides simple authentication scheme where your
// secret API key MUST be specified in every request's Authorization header:
//
//	Authorization: Uploadcare.Simple public_key:secret_key
func SimpleAuth(creds APICreds, req *http.Request) {
	setHeader(req, authHeaderKey, simpleAuthParam(creds))
}

// SignBasedAuth provides SHA1 signature based authentication scheme where
// your secret API key is used to derive signature but is not included in
// the request itself. Authorization header looks like this:
//
//	Authorization: Uploadcare public_key:signature
//
// For more info on how SHA1 signature is constructed see
// https://uploadcare.com/docs/api_reference/rest/requests_auth/
func SignBasedAuth(creds APICreds, req *http.Request) {
	authParam := signBasedAuthParam(creds, req, time.Now())
	setHeader(req, authHeaderKey, authParam)
}

func setHeader(req *http.Request, key, val string) { req.Header.Set(key, val) }

func simpleAuthParam(creds APICreds) string {
	val := fmt.Sprintf(
		"%s %s:%s",
		simpleAuthScheme,
		creds.PublicKey,
		creds.SecretKey,
	)
	log.Debugf("preparing simple auth param: %s", val)
	return val
}

func signBasedAuthParam(creds APICreds, req *http.Request, t time.Time) string {
	bodyData := new(bytes.Buffer)
	var bodyReader io.Reader
	bodyReader = req.Body
	if bodyReader == nil {
		bodyReader = strings.NewReader("")
	}
	io.Copy(bodyData, bodyReader)
	bodyHash := fmt.Sprintf("%x", md5.Sum(bodyData.Bytes()))

	uri := req.URL.Path
	if rq := req.URL.RawQuery; rq != "" {
		uri += "?" + rq
	}

	signData := new(bytes.Buffer)
	signData.WriteString(req.Method)
	signData.WriteRune('\n')
	signData.WriteString(bodyHash)
	signData.WriteRune('\n')
	signData.WriteString(req.Header.Get("Content-Type"))
	signData.WriteRune('\n')
	signData.WriteString(t.In(dateHeaderLocation).Format(dateHeaderFormat))
	signData.WriteRune('\n')
	signData.WriteString(uri)

	h := hmac.New(sha1.New, []byte(creds.SecretKey))
	h.Write(signData.Bytes())
	signature := hex.EncodeToString(h.Sum(nil))

	val := fmt.Sprintf(
		"%s %s:%s",
		signBasedAuthScheme,
		creds.PublicKey,
		signature,
	)

	log.Debugf("preparing sign based auth param: %s", val)
	return val
}
