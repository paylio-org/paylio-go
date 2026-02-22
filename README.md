# Paylio Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/paylio-org/paylio-go.svg)](https://pkg.go.dev/github.com/paylio-org/paylio-go)
[![CI](https://github.com/paylio-org/paylio-go/actions/workflows/ci.yml/badge.svg)](https://github.com/paylio-org/paylio-go/actions/workflows/ci.yml)

The Paylio Go SDK provides convenient access to the Paylio API from applications written in Go.

## Documentation

See the [Paylio API docs](https://paylio.pro/docs).

## Requirements

- Go 1.22+

## Installation

```bash
go get github.com/paylio-org/paylio-go
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	paylio "github.com/paylio-org/paylio-go"
)

func main() {
	client, err := paylio.NewClient("sk_live_xxx")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Retrieve current subscription
	sub, err := client.Subscription.Retrieve(ctx, "user_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Status: %s, Plan: %s\n", sub.Status, sub.Plan.Name)
}
```

### List subscription history

```go
list, err := client.Subscription.List(ctx, "user_123", &paylio.ListOptions{
    Page:     1,
    PageSize: 10,
})
if err != nil {
    log.Fatal(err)
}
for _, item := range list.Items {
    fmt.Printf("%s — %s (%s)\n", item.ID, item.PlanName, item.Status)
}
fmt.Println("Has more:", list.HasMore())
```

### Cancel a subscription

```go
// Cancel at end of billing period (safe default)
result, err := client.Subscription.Cancel(ctx, "sub_uuid", nil)

// Cancel immediately
result, err := client.Subscription.Cancel(ctx, "sub_uuid", &paylio.CancelOptions{
    CancelNow: true,
})
```

### Configuration

```go
// Custom base URL and timeout
client, err := paylio.NewClient("sk_live_xxx",
    paylio.WithBaseURL("https://custom-api.example.com/v1"),
    paylio.WithTimeout(60 * time.Second),
)

// Custom HTTP client
client, err := paylio.NewClient("sk_live_xxx",
    paylio.WithHTTPClient(&http.Client{
        Transport: customTransport,
    }),
)
```

### Error handling

```go
sub, err := client.Subscription.Retrieve(ctx, "user_123")
if err != nil {
    var authErr *paylio.AuthenticationError
    var notFoundErr *paylio.NotFoundError
    var rateLimitErr *paylio.RateLimitError
    var paylioErr *paylio.PaylioError

    switch {
    case errors.As(err, &authErr):
        fmt.Println("Invalid API key:", authErr.Message)
    case errors.As(err, &notFoundErr):
        fmt.Println("Not found:", notFoundErr.Message)
    case errors.As(err, &rateLimitErr):
        fmt.Println("Rate limited, try again later")
    case errors.As(err, &paylioErr):
        fmt.Printf("API error %d: %s\n", paylioErr.HTTPStatus, paylioErr.Message)
    default:
        fmt.Println("Unexpected error:", err)
    }
}
```

## Error types

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| `AuthenticationError` | 401 | Invalid or missing API key |
| `InvalidRequestError` | 400 | Bad request parameters |
| `NotFoundError` | 404 | Resource not found |
| `RateLimitError` | 429 | Rate limit exceeded |
| `APIError` | 5xx | Server error |
| `APIConnectionError` | — | Network or connection failure |

All error types embed `*PaylioError` and work with `errors.As`.

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
