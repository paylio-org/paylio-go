package paylio

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewClientSuccess(t *testing.T) {
	client, err := NewClient("sk_test_key")
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientEmptyKeyReturnsAuthError(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthenticationError, got %T", err)
	}
	if authErr.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestNewClientSubscriptionServiceNotNil(t *testing.T) {
	client, err := NewClient("sk_test")
	if err != nil {
		t.Fatal(err)
	}
	if client.Subscription == nil {
		t.Error("Subscription service is nil")
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	client, err := NewClient("sk_test", WithBaseURL("https://custom.api.com/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	client, err := NewClient("sk_test", WithTimeout(60*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	client, err := NewClient("sk_test", WithHTTPClient(custom))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientMultipleOptions(t *testing.T) {
	client, err := NewClient("sk_test",
		WithBaseURL("https://custom.api.com/v1"),
		WithTimeout(60*time.Second),
		WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestClientCloseIsCallable(t *testing.T) {
	client, err := NewClient("sk_test")
	if err != nil {
		t.Fatal(err)
	}
	client.Close() // should not panic
}

func TestClientCloseMultipleTimes(t *testing.T) {
	client, err := NewClient("sk_test")
	if err != nil {
		t.Fatal(err)
	}
	client.Close()
	client.Close() // second call should not panic
}
