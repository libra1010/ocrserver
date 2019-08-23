package koloCore

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type requestUtil uint

type Method string

const (
	GET    Method = http.MethodGet
	PUT    Method = http.MethodPut
	POST   Method = http.MethodPost
	DELETE Method = http.MethodDelete
	OPTION Method = http.MethodOptions

	ContentType_FORM  string = "application/x-www-form-urlencoded"
	ContentType_MULTI string = "multipart/form-data"
	ContentType_TEXT  string = "text/plain"
	ContentType_JSON  string = "application/json"
	ContentType_HTML  string = "text/html"
	ContentType_XML   string = "application/xml"
)

const RequestUtil requestUtil = 1

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client = &http.Client{Transport: tr}

func (this requestUtil) Request(url string, method Method, contentType string, headers map[string]string, body string) ([]byte, error) {
	req, err := getRequest(url, string(method), contentType, headers, body)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(rsp.Body)

	return bytes, err
}

func (this requestUtil) Post(url string, contentType string, headers map[string]string, body string) ([]byte, error) {
	return this.Request(url, POST, contentType, headers, body)
}

func (this requestUtil) Get(url string, contentType string, headers map[string]string, body string) ([]byte, error) {
	return this.Request(url, GET, contentType, headers, body)
}

func (this requestUtil) Put(url string, contentType string, headers map[string]string, body string) ([]byte, error) {
	return this.Request(url, PUT, contentType, headers, body)
}

func (this requestUtil) RequestBySign(uri string, method Method, contentType string, body string, ident SignIdentity, signKey string) ([]byte, error) {

	headers := make(map[string]string)
	headers["X-UUID"] = ident.UUID
	headers["X-TS"] = strconv.FormatInt(ident.TS, 10)
	headers["X-DEBUG"] = strconv.FormatBool(ident.Debug)
	headers["X-PLATFORM"] = ident.Platform
	headers["X-SKIP_SIGN"] = strconv.FormatBool(ident.SkipSign)
	headers["X-LANGUAGE"] = ident.Language
	headers["User-Agent"] = ident.Agent
	headers["VERSION"] = ident.Version
	ident.Url, _ = url.Parse(uri)

	sign := SignUtil.Sign(ident, signKey)
	headers["X-SIGN"] = sign

	return this.Request(uri, method, contentType, headers, body)
}

func (this requestUtil) PostBySign(url string, contentType string, body string, ident SignIdentity, signKey string) ([]byte, error) {
	return this.RequestBySign(url, POST, contentType, body, ident, signKey)
}

func (this requestUtil) GetBySign(url string, contentType string, body string, ident SignIdentity, signKey string) ([]byte, error) {
	return this.RequestBySign(url, GET, contentType, body, ident, signKey)
}

func (this requestUtil) PutBySign(url string, contentType string, body string, ident SignIdentity, signKey string) ([]byte, error) {
	return this.RequestBySign(url, PUT, contentType, body, ident, signKey)
}

func getRequest(url string, method string, contentType string, headers map[string]string, body string) (*http.Request, error) {
	var reader io.Reader = nil

	body = strings.TrimSpace(body)
	if body != "" && !strings.EqualFold(method, "get") {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if headers != nil && len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	req.Close = true

	return req, err
}
