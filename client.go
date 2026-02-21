package paylio

import (
	"net/http"
	"time"
)

// Client is the entry point for the Paylio SDK.
type Client struct {
	// Subscription provides access to subscription operations.
	Subscription *SubscriptionService

	hc *httpClient
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// WithBaseURL sets a custom base URL for API requests.
func WithBaseURL(url string) Option {
	return func(c *clientConfig) { c.baseURL = url }
}

// WithTimeout sets a custom request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) { c.timeout = timeout }
}

// WithHTTPClient sets a custom net/http client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) { c.httpClient = client }
}

// NewClient creates a new Paylio SDK client.
// Returns an AuthenticationError if apiKey is empty.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, NewAuthenticationError(ErrorParams{
			Message: "No API key provided. Set your API key when creating the client: paylio.NewClient(\"sk_live_xxx\")",
		})
	}

	cfg := &clientConfig{
		baseURL:    DefaultBaseURL,
		timeout:    DefaultTimeout,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	hc := newHTTPClient(apiKey, cfg.baseURL, cfg.timeout, cfg.httpClient)
	return &Client{
		Subscription: newSubscriptionService(hc),
		hc:           hc,
	}, nil
}

// Close releases resources held by the client.
func (c *Client) Close() {
	c.hc.close()
}
