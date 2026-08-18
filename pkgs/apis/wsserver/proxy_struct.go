package wsserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
)

type ProxyRequest struct {
	TXID string

	Method string
	URL    *url.URL

	Proto string // "HTTP/1.0"

	Header http.Header

	Body []byte

	ContentLength int64

	Host string

	Form url.Values

	PostForm url.Values

	MultipartForm *multipart.Form

	RemoteAddr string

	RequestURI string
}

func NewProxyRequest(req *http.Request) (*ProxyRequest, error) {
	var body []byte
	var err error
	if req.Body != nil {
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			logs.ErrorContextf(req.Context(), "[httpProxy] read body failed, %s", err)
			return nil, err
		}
	}
	prouter := &ProxyRequest{
		TXID: types.GenerateID(),

		Method:        req.Method,
		URL:           req.URL,
		Proto:         req.Proto,
		Header:        req.Header,
		Body:          body,
		ContentLength: req.ContentLength,

		Host:          req.Host,
		Form:          req.Form,
		PostForm:      req.PostForm,
		MultipartForm: req.MultipartForm,
		RemoteAddr:    req.RemoteAddr,
		RequestURI:    req.RequestURI,
	}
	return prouter, nil
}

func (pr ProxyRequest) ToHTTPRequest() (*http.Request, error) {
	req := &http.Request{
		Method:        pr.Method,
		URL:           pr.URL,
		Proto:         pr.Proto,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(bytes.NewReader(pr.Body)),
		ContentLength: pr.ContentLength,
	}

	for k, v := range pr.Header {
		req.Header[k] = v
	}

	return req, nil
}

type ProxyResponse struct {
	TXID string

	ErrMessage    string
	Status        string // e.g. "200 OK"
	StatusCode    int    // e.g. 200
	Header        http.Header
	Body          []byte
	ContentLength int64
}

func NewProxyResponse(ctx context.Context, txid string, resp *http.Response, err error) *ProxyResponse {
	pr := &ProxyResponse{
		TXID: txid,
	}
	if err != nil {
		pr.ErrMessage = err.Error()
		return pr
	}

	if resp.Body != nil {
		pr.Body, err = io.ReadAll(resp.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "[httpProxy] read body failed, %s", err)
			pr.ErrMessage = err.Error()
			return pr
		}
	}

	pr.Status = resp.Status
	pr.StatusCode = resp.StatusCode
	pr.Header = resp.Header
	pr.ContentLength = resp.ContentLength
	return pr
}

func (pr ProxyResponse) ToHTTPResponse() (*http.Response, error) {
	if pr.StatusCode == 0 {
		return nil, fmt.Errorf(pr.ErrMessage)
	}
	resp := &http.Response{
		Status:        pr.Status,
		StatusCode:    pr.StatusCode,
		Header:        pr.Header,
		Body:          io.NopCloser(bytes.NewReader(pr.Body)),
		ContentLength: pr.ContentLength,
	}

	return resp, nil
}
