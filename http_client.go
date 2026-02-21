package paylio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default base URL for the Paylio API.
	DefaultBaseURL = "https://api.paylio.pro/flying/v1"

	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 30 * time.Second
)

type httpClient struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	client  *http.Client
}

type requestOptions struct {
	Params   map[string]string
	JSONBody map[string]any
}

func newHTTPClient(apiKey, baseURL string, timeout time.Duration, client *http.Client) *httpClient {
	return &httpClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: timeout,
		client:  client,
	}
}

func (hc *httpClient) request(ctx context.Context, method, path string, opts *requestOptions) (map[string]any, error) {
	fullURL := hc.baseURL + path

	if opts != nil && opts.Params != nil {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, NewAPIConnectionError(ErrorParams{Message: fmt.Sprintf("failed to parse URL: %v", err)})
		}
		q := u.Query()
		for k, v := range opts.Params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	var body io.Reader
	if opts != nil && opts.JSONBody != nil {
		b, err := json.Marshal(opts.JSONBody)
		if err != nil {
			return nil, NewAPIConnectionError(ErrorParams{Message: fmt.Sprintf("failed to marshal body: %v", err)})
		}
		body = bytes.NewReader(b)
	}

	ctx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, NewAPIConnectionError(ErrorParams{Message: fmt.Sprintf("failed to create request: %v", err)})
	}

	req.Header.Set("X-API-Key", hc.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "paylio-go/"+Version)
	req.Header.Set("X-SDK-Source", "go")

	resp, err := hc.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewAPIConnectionError(ErrorParams{Message: "Request timed out"})
		}
		return nil, NewAPIConnectionError(ErrorParams{Message: fmt.Sprintf("Connection error: %v", err)})
	}
	defer resp.Body.Close()

	return hc.handleResponse(resp)
}

func (hc *httpClient) handleResponse(resp *http.Response) (map[string]any, error) {
	httpStatus := resp.StatusCode
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAPIConnectionError(ErrorParams{Message: fmt.Sprintf("failed to read response body: %v", err)})
	}
	httpBody := string(bodyBytes)

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	var jsonBody map[string]any
	if err := json.Unmarshal(bodyBytes, &jsonBody); err != nil {
		jsonBody = nil
	}

	if httpStatus >= 200 && httpStatus < 300 {
		if jsonBody == nil {
			return nil, NewAPIError(ErrorParams{
				Message:    "Invalid JSON in response body",
				HTTPStatus: httpStatus,
				HTTPBody:   httpBody,
			})
		}
		return jsonBody, nil
	}

	errorCode := ""
	errorMessage := httpBody

	if jsonBody != nil {
		if errField, ok := jsonBody["error"]; ok {
			switch e := errField.(type) {
			case map[string]any:
				if code, ok := e["code"].(string); ok {
					errorCode = code
				}
				if msg, ok := e["message"].(string); ok {
					errorMessage = msg
				}
			case string:
				errorMessage = e
			}
		} else if detail, ok := jsonBody["detail"].(string); ok {
			errorMessage = detail
		}
	}

	params := ErrorParams{
		Message:    errorMessage,
		HTTPStatus: httpStatus,
		HTTPBody:   httpBody,
		JSONBody:   jsonBody,
		Headers:    headers,
		Code:       errorCode,
	}

	return nil, errorClassForStatus(httpStatus, params)
}

func (hc *httpClient) close() {
	hc.client.CloseIdleConnections()
}
