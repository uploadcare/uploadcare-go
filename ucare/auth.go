package ucare

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type restAPIAuthFunc func(APICreds, *http.Request)

func simpleRESTAPIAuth(creds APICreds, req *http.Request) {
	setHeader(req, authHeaderKey, simpleRESTAPIAuthParam(creds))
}

func signBasedRESTAPIAuth(creds APICreds, req *http.Request) {
	authParam := signBasedRESTAPIAuthParam(creds, req)
	setHeader(req, authHeaderKey, authParam)
}

func setHeader(req *http.Request, key, val string) { req.Header.Set(key, val) }

func simpleRESTAPIAuthParam(creds APICreds) string {
	val := fmt.Sprintf(
		"%s %s:%s",
		simpleAuthScheme,
		creds.PublicKey,
		creds.SecretKey,
	)
	log.Debugf("preparing simple auth param: %s", val)
	return val
}

func signBasedRESTAPIAuthParam(creds APICreds, req *http.Request) string {
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
	signData.WriteString(req.Header.Get("Date"))
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

// UploadAPIAuthFunc is for internal use and should not be used by users
type UploadAPIAuthFunc func() (pubkey string, sign *string, exp *int64)

func simpleUploadAPIAuthFunc(creds APICreds) UploadAPIAuthFunc {
	return func() (string, *string, *int64) {
		return creds.PublicKey, nil, nil
	}
}

func signBasedUploadAPIAuthFunc(creds APICreds) UploadAPIAuthFunc {
	return func() (string, *string, *int64) {
		exp := time.Now().Add(30 * time.Minute).Unix()
		sign := signBasedUploadAPIAuthParam(creds.SecretKey, exp)
		return creds.PublicKey, &sign, &exp
	}
}

func signBasedUploadAPIAuthParam(secret string, exp int64) string {
	h := md5.New()
	h.Write([]byte(secret + strconv.FormatInt(exp, 10)))
	return hex.EncodeToString(h.Sum(nil))
}
