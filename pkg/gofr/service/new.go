package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type httpService struct {
	*http.Client

	url string
}

func (h *httpService) Get(ctx context.Context, path string, params map[string]interface{}) (*http.Response, error) {
	return h.GetWithHeaders(ctx, path, params, nil)
}
func (h *httpService) GetWithHeaders(ctx context.Context, path string, params map[string]interface{}, headers map[string]string) (*http.Response, error) {
	uri := h.url + "/" + path

	tr := otel.Tracer("gofr-http-client-get")

	ctx, span := tr.Start(ctx, uri)
	defer span.End()

	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	req, _ := http.NewRequestWithContext(ctx, "GET", uri, http.NoBody)

	addHeadersToRequest(req, headers)
	encodeQueryParameters(req, params)

	return h.Do(req)
}

func (h *httpService) Post(ctx context.Context, path string, params map[string]interface{}, body []byte) (*http.Response, error) {
	return h.PostWithHeaders(ctx, path, params, body, nil)
}

func (h *httpService) PostWithHeaders(ctx context.Context, path string, params map[string]interface{}, body []byte, headers map[string]string) (*http.Response, error) {
	uri := h.url + "/" + path

	if path == "" {
		uri = h.url
	}

	tr := otel.Tracer("gofr-http-client-post")

	ctx, span := tr.Start(ctx, uri)
	defer span.End()

	ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBuffer(body))

	addHeadersToRequest(req, headers)
	encodeQueryParameters(req, params)

	return h.Do(req)
}

type HTTP interface {
	Get(ctx context.Context, api string, params map[string]interface{}) (*http.Response, error)
	GetWithHeaders(ctx context.Context, api string, params map[string]interface{}, headers map[string]string) (*http.Response, error)
	Post(ctx context.Context, api string, params map[string]interface{}, body []byte) (*http.Response, error)
	PostWithHeaders(ctx context.Context, api string, params map[string]interface{}, body []byte, header map[string]string) (*http.Response, error)
}

func NewHTTPService(serviceAddress string) HTTP {
	return &httpService{
		Client: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		url:    serviceAddress,
	}
}

func encodeQueryParameters(req *http.Request, params map[string]interface{}) {
	q := req.URL.Query()

	for k, v := range params {
		switch vt := v.(type) {
		case []string:
			for _, val := range vt {
				q.Set(k, val)
			}
		default:
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}

	req.URL.RawQuery = q.Encode()
}

func addHeadersToRequest(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}
