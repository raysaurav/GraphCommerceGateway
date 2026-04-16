package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-resty/resty/v2"
)

type HttpClientInterface interface {
	Get(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string) (*resty.Response, error)
	Post(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error)
	Patch(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error)
	Put(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error)
	Delete(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string) (*resty.Response, error)
	DeleteWithReqBody(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error)
}

type contextKey string

const TimeoutKey contextKey = "timeout"

// SetRequestParams sets headers, query parameters, and path parameters for a request
func SetRequestParams(req *resty.Request, headers, queryParams, pathParams map[string]string) {
	// Set headers
	for key, value := range headers {
		req.SetHeader(key, value)
	}

	// Set query parameters
	for key, value := range queryParams {
		req.SetQueryParam(key, value)
	}

	// Set path parameters
	for key, value := range pathParams {
		req.SetPathParam(key, value)
	}
}

// manual retry helper
// simple one-retry helper for resiliency path
func (hc *HttpClient) retryProxy(
	ctx context.Context,
	do func(context.Context) (*resty.Response, error),
) (*resty.Response, error) {
	// first attempt
	resp, err := do(ctx)
	if ShouldRetry(resp, err) {
		// wait 100s, honoring context
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		// single retry
		return do(ctx)
	}
	return resp, err
}

// doRequest builds and executes an HTTP request using the Resty client.
// It sets request parameters, handles retries for proxied URLs via resilience logic,
// logs execution time, and reports metrics. It supports standard HTTP methods.
//
// Returns the HTTP response or an error if the request fails.

func (hc *HttpClient) doRequest(
	ctx context.Context,
	method string,
	url string,
	headers map[string]string,
	queryParams map[string]string,
	pathParams map[string]string,
	body interface{},
	graphOperationName string,
) (*resty.Response, error) {
	timeout := GetTimeoutFromContext(ctx)
	hc.RestyClient.SetTimeout(timeout)
	hc.RestyClient.SetDebug(true)

	req := hc.RestyClient.R().SetContext(ctx)
	if body != nil {
		req.SetBody(body)
	}
	SetRequestParams(req, headers, queryParams, pathParams)

	var resp *resty.Response
	var err error
	switch method {
	case http.MethodGet:
		resp, err = req.Get(url)
	case http.MethodPost:
		resp, err = req.Post(url)
	case http.MethodPut:
		resp, err = req.Put(url)
	case http.MethodPatch:
		resp, err = req.Patch(url)
	case http.MethodDelete:
		resp, err = req.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
	return resp, err
}

// prepareAndDoRequest enriches the request with observability headers,
// rewrites the URL for performance, and delegates execution to doRequest.
// If a circuit breaker is triggered, it records appropriate metrics.
//
// Returns the HTTP response or an error if the request fails or the circuit is open.

func (hc *HttpClient) prepareAndDoRequest(
	ctx context.Context,
	method string,
	url string,
	headers map[string]string,
	queryParams map[string]string,
	pathParams map[string]string,
	body interface{},
) (*resty.Response, error) {

	url, headers, body, pathParams, graphOperationName := addObservabilityHeaders(ctx, url, headers, pathParams, body)

	resp, err := hc.doRequest(ctx, method, url, headers, queryParams, pathParams, body, graphOperationName)

	if err != nil && strings.Contains(err.Error(), "circuit breaker is open") {
		fmt.Printf("Circuit breaker triggered for %s request to %s: %v", method, url, err)
	}
	return resp, err
}

// Get performs a GET request
func (hc *HttpClient) Get(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodGet, url, headers, queryParams, pathParams, nil)
}

// Post performs a POST request with optional headers, query parameters, path parameters, and body
func (hc *HttpClient) Post(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodPost, url, headers, queryParams, pathParams, body)
}

// Patch performs a PATCH request with optional headers, query parameters, path parameters, and body
func (hc *HttpClient) Patch(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodPatch, url, headers, queryParams, pathParams, body)
}

// Put performs a PUT request with optional headers, query parameters, path parameters, and body
func (hc *HttpClient) Put(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodPut, url, headers, queryParams, pathParams, body)
}

// Delete performs a DELETE request with optional headers, query parameters, and path parameters
func (hc *HttpClient) Delete(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodDelete, url, headers, queryParams, pathParams, nil)
}

// DeleteWithReqBody performs a DELETE request with optional headers, query parameters, path parameters, and a request body
func (hc *HttpClient) DeleteWithReqBody(ctx context.Context, url string, headers map[string]string, queryParams, pathParams map[string]string, body interface{}) (*resty.Response, error) {
	return hc.prepareAndDoRequest(ctx, http.MethodDelete, url, headers, queryParams, pathParams, body)
}

// GetTimeoutFromContext retrieves the timeout value from the context.
// If no timeout is found, it returns a default timeout.
func GetTimeoutFromContext(ctx context.Context) time.Duration {
	// Attempt to fetch the timeout from the context
	timeout, ok := ctx.Value(TimeoutKey).(time.Duration)
	if !ok {
		// Log and return the default timeout if not found
		timeout = 30 * time.Second
		return timeout
	}
	return timeout
}

func addObservabilityHeaders(
	ctx context.Context,
	rawURL string,
	headers, path map[string]string,
	body interface{},
) (retURL string, retHeaders map[string]string, retBody interface{}, retPath map[string]string, graphOperationName string) {

	// Initialize return values to inputs (fail-safe fallback)
	retURL = rawURL
	retHeaders = headers
	retBody = body
	retPath = path
	graphOperationName = ""

	// Panic recovery to return inputs in case of crash
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in addObservabilityHeaders: %v", r)
		}
	}()

	// If headers is nil, create a new one
	if retHeaders == nil {
		retHeaders = make(map[string]string)
	}

	// Get GraphQL field context
	if fc := graphql.GetFieldContext(ctx); fc != nil {
		retHeaders["x-graphql-method"] = fc.Field.Name
		graphOperationName = fc.Field.Name
	}

	return retURL, retHeaders, retBody, retPath, graphOperationName
}
