package paylio

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConstants(t *testing.T) {
	if DefaultBaseURL != "https://api.paylio.pro/flying/v1" {
		t.Errorf("DefaultBaseURL = %q", DefaultBaseURL)
	}
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v", DefaultTimeout)
	}
}

func TestHTTPClientSendsCorrectHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != "sk_test_key" {
			t.Errorf("X-API-Key = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "paylio-go/"+Version {
			t.Errorf("User-Agent = %q", got)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test_key", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPClientSuccessReturnsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id": "sub_1", "status": "active"}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	data, err := hc.request(context.Background(), "GET", "/sub", nil)
	if err != nil {
		t.Fatal(err)
	}
	if data["id"] != "sub_1" {
		t.Errorf("id = %v", data["id"])
	}
	if data["status"] != "active" {
		t.Errorf("status = %v", data["status"])
	}
}

func TestHTTPClientNonJSONSuccessReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/bad", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
}

func TestHTTPClientErrorStatusMapping(t *testing.T) {
	tests := []struct {
		status   int
		errCheck func(error) bool
		name     string
	}{
		{401, func(e error) bool { var v *AuthenticationError; return errors.As(e, &v) }, "401->AuthenticationError"},
		{400, func(e error) bool { var v *InvalidRequestError; return errors.As(e, &v) }, "400->InvalidRequestError"},
		{404, func(e error) bool { var v *NotFoundError; return errors.As(e, &v) }, "404->NotFoundError"},
		{429, func(e error) bool { var v *RateLimitError; return errors.As(e, &v) }, "429->RateLimitError"},
		{500, func(e error) bool { var v *APIError; return errors.As(e, &v) }, "500->APIError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error": "test"}`))
			}))
			defer srv.Close()

			hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
			_, err := hc.request(context.Background(), "GET", "/err", nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if !tt.errCheck(err) {
				t.Errorf("wrong error type for status %d: %T", tt.status, err)
			}
		})
	}
}

func TestHTTPClientErrorFormatV1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error": {"code": "invalid_param", "message": "bad field"}}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/v1", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Message != "bad field" {
		t.Errorf("Message = %q", pe.Message)
	}
	if pe.Code != "invalid_param" {
		t.Errorf("Code = %q", pe.Code)
	}
}

func TestHTTPClientErrorFormatLegacy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error": "legacy error message"}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/legacy", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Message != "legacy error message" {
		t.Errorf("Message = %q", pe.Message)
	}
}

func TestHTTPClientErrorFormatFastAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"detail": "fastapi error"}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/fastapi", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Message != "fastapi error" {
		t.Errorf("Message = %q", pe.Message)
	}
}

func TestHTTPClientErrorNonJSONFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`Internal Server Error`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/raw", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Message != "Internal Server Error" {
		t.Errorf("Message = %q", pe.Message)
	}
	if pe.HTTPBody != "Internal Server Error" {
		t.Errorf("HTTPBody = %q", pe.HTTPBody)
	}
}

func TestHTTPClientErrorPreservesHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "req_abc")
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"error": "fail"}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/headers", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Headers["X-Request-Id"] != "req_abc" {
		t.Errorf("Headers = %v", pe.Headers)
	}
}

func TestHTTPClientErrorPreservesJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error": "test"}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/json", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.JSONBody == nil {
		t.Fatal("JSONBody should not be nil")
	}
	if pe.JSONBody["error"] != "test" {
		t.Errorf("JSONBody = %v", pe.JSONBody)
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 50*time.Millisecond, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/slow", nil)

	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

func TestHTTPClientConnectionError(t *testing.T) {
	// Connect to a port that's not listening
	hc := newHTTPClient("sk_test", "http://127.0.0.1:1", 5*time.Second, &http.Client{})
	_, err := hc.request(context.Background(), "GET", "/fail", nil)

	var connErr *APIConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *APIConnectionError, got %T: %v", err, err)
	}
}

func TestHTTPClientGETWithParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("page_size = %q", r.URL.Query().Get("page_size"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/list", &requestOptions{
		Params: map[string]string{"page": "2", "page_size": "10"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPClientPOSTWithJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed["cancel_at_period_end"] != true {
			t.Errorf("body = %v", parsed)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "POST", "/cancel", &requestOptions{
		JSONBody: map[string]any{"cancel_at_period_end": true},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPClientCustomBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/custom/path" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL+"/custom/", 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/path", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPClientCloseIsCallable(t *testing.T) {
	hc := newHTTPClient("sk_test", "http://localhost", 10*time.Second, &http.Client{})
	hc.close() // should not panic
}

func TestHTTPClientV1ErrorNonStringCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error": {"code": 123, "message": "msg"}}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/v1", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	if pe.Code != "" {
		t.Errorf("Code should be empty for non-string, got %q", pe.Code)
	}
	if pe.Message != "msg" {
		t.Errorf("Message = %q", pe.Message)
	}
}

func TestHTTPClientV1ErrorNonStringMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		body := `{"error": {"code": "err", "message": 999}}`
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/v1", nil)

	var pe *PaylioError
	if !errors.As(err, &pe) {
		t.Fatal("expected PaylioError")
	}
	// When message is non-string, falls back to httpBody
	if pe.Message != `{"error": {"code": "err", "message": 999}}` {
		t.Errorf("Message = %q", pe.Message)
	}
}

func TestHTTPClientNoBodyOnGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("GET should not have body, got %q", string(body))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	_, err := hc.request(context.Background(), "GET", "/no-body", nil)
	if err != nil {
		t.Fatal(err)
	}
}
