package paylio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"testing"
	"time"
)

func TestUnmarshalToMarshalFailure(t *testing.T) {
	// math.NaN() cannot be marshaled to JSON
	data := map[string]any{"value": math.NaN()}
	_, err := unmarshalTo[Subscription](data)
	if err == nil {
		t.Fatal("expected error for unmarshalable data")
	}
}

func TestUnmarshalToUnmarshalFailure(t *testing.T) {
	// "plan" expects a struct but gets a string to trigger unmarshal error.
	data := map[string]any{"plan": "not-a-plan-object"}
	_, err := unmarshalTo[Subscription](data)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestHTTPClientBodyMarshalError(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	_, err := hc.request(context.Background(), "POST", "/test", &requestOptions{
		JSONBody: map[string]any{"bad": math.NaN()},
	})
	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

func TestHTTPClientInvalidMethodError(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	_, err := hc.request(context.Background(), "BAD METHOD", "/test", nil)
	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

// errReader is a reader that always returns an error.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }
func (errReader) Close() error             { return nil }

func TestHandleResponseBodyReadError(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       io.NopCloser(errReader{}),
	}
	_, err := hc.handleResponse(resp)
	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

func TestHTTPClientURLParseError(t *testing.T) {
	// Use a base URL with control characters to trigger url.Parse error
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	hc.baseURL = string([]byte{0x7f}) // DEL character causes parse failure
	_, err := hc.request(context.Background(), "GET", "/test", &requestOptions{
		Params: map[string]string{"key": "val"},
	})
	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

func TestHTTPClientErrorFormatUnrecognizedJSON(t *testing.T) {
	// JSON body without error or detail keys — falls back to raw body
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	body := `{"unknown_key": "value"}`
	resp := &http.Response{
		StatusCode: 500,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
	_, err := hc.handleResponse(resp)
	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Message != body {
		t.Errorf("Message = %q, expected raw body", pe.Message)
	}
}

func TestRetrieveAPIErrorPropagation(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://127.0.0.1:1", 5*time.Second, &http.Client{})
	svc := newSubscriptionService(hc)
	_, err := svc.Retrieve(context.Background(), "user_1")
	if err == nil {
		t.Fatal("expected error")
	}
	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T", err)
	}
}

func TestListAPIErrorPropagation(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://127.0.0.1:1", 5*time.Second, &http.Client{})
	svc := newSubscriptionService(hc)
	_, err := svc.List(context.Background(), "user_1", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelAPIErrorPropagation(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://127.0.0.1:1", 5*time.Second, &http.Client{})
	svc := newSubscriptionService(hc)
	_, err := svc.Cancel(context.Background(), "sub_1", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
