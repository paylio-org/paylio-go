package paylio

// ErrorParams holds the parameters for constructing a PaylioError.
type ErrorParams struct {
	Message    string
	HTTPStatus int
	HTTPBody   string
	JSONBody   map[string]any
	Headers    map[string]string
	Code       string
}

// PaylioError is the base error type for all Paylio SDK errors.
type PaylioError struct {
	Message    string
	HTTPStatus int
	HTTPBody   string
	JSONBody   map[string]any
	Headers    map[string]string
	Code       string
}

func (e *PaylioError) Error() string { return e.Message }

func newPaylioError(p ErrorParams) *PaylioError {
	return &PaylioError{
		Message:    p.Message,
		HTTPStatus: p.HTTPStatus,
		HTTPBody:   p.HTTPBody,
		JSONBody:   p.JSONBody,
		Headers:    p.Headers,
		Code:       p.Code,
	}
}

// APIError indicates a server error or unexpected HTTP status.
type APIError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *APIError) Unwrap() error { return e.PaylioError }

// NewAPIError creates an APIError from the given params.
func NewAPIError(p ErrorParams) *APIError {
	return &APIError{newPaylioError(p)}
}

// AuthenticationError indicates an invalid or missing API key (HTTP 401).
type AuthenticationError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *AuthenticationError) Unwrap() error { return e.PaylioError }

// NewAuthenticationError creates an AuthenticationError from the given params.
func NewAuthenticationError(p ErrorParams) *AuthenticationError {
	return &AuthenticationError{newPaylioError(p)}
}

// InvalidRequestError indicates bad request parameters (HTTP 400).
type InvalidRequestError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *InvalidRequestError) Unwrap() error { return e.PaylioError }

// NewInvalidRequestError creates an InvalidRequestError from the given params.
func NewInvalidRequestError(p ErrorParams) *InvalidRequestError {
	return &InvalidRequestError{newPaylioError(p)}
}

// NotFoundError indicates a resource was not found (HTTP 404).
type NotFoundError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *NotFoundError) Unwrap() error { return e.PaylioError }

// NewNotFoundError creates a NotFoundError from the given params.
func NewNotFoundError(p ErrorParams) *NotFoundError {
	return &NotFoundError{newPaylioError(p)}
}

// RateLimitError indicates rate limit exceeded (HTTP 429).
type RateLimitError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *RateLimitError) Unwrap() error { return e.PaylioError }

// NewRateLimitError creates a RateLimitError from the given params.
func NewRateLimitError(p ErrorParams) *RateLimitError {
	return &RateLimitError{newPaylioError(p)}
}

// APIConnectionError indicates a network failure or timeout.
type APIConnectionError struct{ *PaylioError }

// Unwrap returns the underlying PaylioError.
func (e *APIConnectionError) Unwrap() error { return e.PaylioError }

// NewAPIConnectionError creates an APIConnectionError from the given params.
func NewAPIConnectionError(p ErrorParams) *APIConnectionError {
	return &APIConnectionError{newPaylioError(p)}
}

// errorClassForStatus returns the appropriate error for the given HTTP status.
func errorClassForStatus(status int, p ErrorParams) error {
	switch status {
	case 401:
		return NewAuthenticationError(p)
	case 400:
		return NewInvalidRequestError(p)
	case 404:
		return NewNotFoundError(p)
	case 429:
		return NewRateLimitError(p)
	default:
		return NewAPIError(p)
	}
}
