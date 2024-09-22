package httpio

import (
	"context"
	"net/http"
)

func New(url string, opts ...RequestOption) *Request {
	r := &Request{requests: opts}

	if url != "" {
		r.With(func(ro *RequestOptions) { ro.URL = url })
	}

	return r
}

type Request struct {
	clients     []ClientOption
	requests    []RequestOption
	middlewares []ResponseMiddleware
}

func (r *Request) With(opts ...RequestOption) *Request {
	r.requests = append(r.requests, opts...)
	return r
}

func (r *Request) Use(middlewares ...ResponseMiddleware) *Request {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *Request) Do(ctx context.Context, process ResponseProcess) (err error) {
	var (
		req    *http.Request
		resp   *http.Response
		client *http.Client
	)

	if req, err = BuildRequest(r.requests...); err != nil {
		return
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	if client, err = BuildClient(r.clients...); err != nil {
		return
	}

	if resp, err = client.Do(req); err != nil {
		return
	}

	return ProcessResponse(resp, process, r.middlewares...)
}
