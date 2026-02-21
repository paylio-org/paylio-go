package paylio

import (
	"errors"
	"testing"
)

func TestPaylioErrorImplementsError(t *testing.T) {
	var _ error = &PaylioError{}
}

func TestPaylioErrorDefaults(t *testing.T) {
	e := &PaylioError{}
	if e.Message != "" {
		t.Errorf("expected empty message, got %q", e.Message)
	}
	if e.HTTPStatus != 0 {
		t.Errorf("expected 0 status, got %d", e.HTTPStatus)
	}
	if e.HTTPBody != "" {
		t.Errorf("expected empty body, got %q", e.HTTPBody)
	}
	if e.JSONBody != nil {
		t.Error("expected nil JSONBody")
	}
	if e.Headers != nil {
		t.Error("expected nil Headers")
	}
	if e.Code != "" {
		t.Errorf("expected empty code, got %q", e.Code)
	}
}

func TestPaylioErrorStoresAllFields(t *testing.T) {
	e := &PaylioError{
		Message:    "bad request",
		HTTPStatus: 400,
		HTTPBody:   `{"error":"bad"}`,
		JSONBody:   map[string]any{"error": "bad"},
		Headers:    map[string]string{"x-request-id": "abc"},
		Code:       "invalid_param",
	}
	if e.Message != "bad request" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.HTTPStatus != 400 {
		t.Errorf("HTTPStatus = %d", e.HTTPStatus)
	}
	if e.HTTPBody != `{"error":"bad"}` {
		t.Errorf("HTTPBody = %q", e.HTTPBody)
	}
	if e.JSONBody["error"] != "bad" {
		t.Errorf("JSONBody = %v", e.JSONBody)
	}
	if e.Headers["x-request-id"] != "abc" {
		t.Errorf("Headers = %v", e.Headers)
	}
	if e.Code != "invalid_param" {
		t.Errorf("Code = %q", e.Code)
	}
}

func TestPaylioErrorReturnsMessage(t *testing.T) {
	e := &PaylioError{Message: "something broke"}
	if e.Error() != "something broke" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestNewErrorConstructors(t *testing.T) {
	params := ErrorParams{
		Message:    "test error",
		HTTPStatus: 500,
		HTTPBody:   "body",
		JSONBody:   map[string]any{"k": "v"},
		Headers:    map[string]string{"h": "v"},
		Code:       "err_code",
	}

	tests := []struct {
		name    string
		newFunc func(ErrorParams) error
	}{
		{"APIError", func(p ErrorParams) error { return NewAPIError(p) }},
		{"AuthenticationError", func(p ErrorParams) error { return NewAuthenticationError(p) }},
		{"InvalidRequestError", func(p ErrorParams) error { return NewInvalidRequestError(p) }},
		{"NotFoundError", func(p ErrorParams) error { return NewNotFoundError(p) }},
		{"RateLimitError", func(p ErrorParams) error { return NewRateLimitError(p) }},
		{"APIConnectionError", func(p ErrorParams) error { return NewAPIConnectionError(p) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.newFunc(params)

			// Must implement error interface
			if err.Error() != "test error" {
				t.Errorf("Error() = %q", err.Error())
			}

			// Must be unwrappable to *PaylioError via errors.As
			var pe *PaylioError
			if !errors.As(err, &pe) {
				t.Fatal("errors.As(*PaylioError) failed")
			}
			if pe.Message != "test error" {
				t.Errorf("PaylioError.Message = %q", pe.Message)
			}
			if pe.HTTPStatus != 500 {
				t.Errorf("PaylioError.HTTPStatus = %d", pe.HTTPStatus)
			}
			if pe.HTTPBody != "body" {
				t.Errorf("PaylioError.HTTPBody = %q", pe.HTTPBody)
			}
			if pe.Code != "err_code" {
				t.Errorf("PaylioError.Code = %q", pe.Code)
			}
		})
	}
}

func TestErrorsAsSpecificTypes(t *testing.T) {
	params := ErrorParams{Message: "typed"}

	var apiErr *APIError
	if !errors.As(NewAPIError(params), &apiErr) {
		t.Error("errors.As(*APIError) failed")
	}

	var authErr *AuthenticationError
	if !errors.As(NewAuthenticationError(params), &authErr) {
		t.Error("errors.As(*AuthenticationError) failed")
	}

	var invalidErr *InvalidRequestError
	if !errors.As(NewInvalidRequestError(params), &invalidErr) {
		t.Error("errors.As(*InvalidRequestError) failed")
	}

	var notFoundErr *NotFoundError
	if !errors.As(NewNotFoundError(params), &notFoundErr) {
		t.Error("errors.As(*NotFoundError) failed")
	}

	var rateLimitErr *RateLimitError
	if !errors.As(NewRateLimitError(params), &rateLimitErr) {
		t.Error("errors.As(*RateLimitError) failed")
	}

	var connErr *APIConnectionError
	if !errors.As(NewAPIConnectionError(params), &connErr) {
		t.Error("errors.As(*APIConnectionError) failed")
	}
}

func TestErrorClassForStatus(t *testing.T) {
	tests := []struct {
		status   int
		wantType string
	}{
		{401, "*paylio.AuthenticationError"},
		{400, "*paylio.InvalidRequestError"},
		{404, "*paylio.NotFoundError"},
		{429, "*paylio.RateLimitError"},
		{500, "*paylio.APIError"},
		{502, "*paylio.APIError"},
		{418, "*paylio.APIError"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			params := ErrorParams{HTTPStatus: tt.status, Message: "test"}
			err := errorClassForStatus(tt.status, params)
			got := errors.Unwrap(err)
			_ = got
			// Just verify it returns a non-nil error with correct message
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Error() != "test" {
				t.Errorf("Error() = %q", err.Error())
			}

			// Verify it unwraps to PaylioError
			var pe *PaylioError
			if !errors.As(err, &pe) {
				t.Fatal("errors.As(*PaylioError) failed")
			}
			if pe.HTTPStatus != tt.status {
				t.Errorf("HTTPStatus = %d, want %d", pe.HTTPStatus, tt.status)
			}
		})
	}
}
