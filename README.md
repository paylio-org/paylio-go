# paylio-go

Go client library for the [Paylio](https://paylio.pro) subscription API.

- **Zero dependencies** — standard library only (`net/http`, `encoding/json`)
- **Type-safe** — strongly typed structs with JSON tags
- **Context-aware** — all methods accept `context.Context` for timeout/cancellation
- **100% test coverage**

## Installation

```bash
go get github.com/paylio-org/paylio-go
```

Requires **Go 1.22** or later.

## Quick Start

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

	// Retrieve a subscription
	sub, err := client.Subscription.Retrieve(ctx, "user_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Status: %s, Plan: %s\n", sub.Status, sub.Plan.Name)
}
```

## Usage

### Retrieve a Subscription

```go
sub, err := client.Subscription.Retrieve(ctx, "user_123")
if err != nil {
    log.Fatal(err)
}
fmt.Println(sub.Status, sub.Plan.Slug, sub.Provider)
```

### List Subscription History

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

### Cancel a Subscription

```go
// Cancel at end of billing period (safe default)
result, err := client.Subscription.Cancel(ctx, "sub_uuid", nil)

// Cancel immediately
result, err := client.Subscription.Cancel(ctx, "sub_uuid", &paylio.CancelOptions{
    CancelNow: true,
})
```

## Error Handling

```go
sub, err := client.Subscription.Retrieve(ctx, "user_123")
if err != nil {
    var authErr *paylio.AuthenticationError
    var notFoundErr *paylio.NotFoundError
    var paylioErr *paylio.PaylioError

    switch {
    case errors.As(err, &authErr):
        fmt.Println("Invalid API key:", authErr.Message)
    case errors.As(err, &notFoundErr):
        fmt.Println("Not found:", notFoundErr.Message)
    case errors.As(err, &paylioErr):
        fmt.Printf("API error %d: %s\n", paylioErr.HTTPStatus, paylioErr.Message)
    default:
        fmt.Println("Unexpected error:", err)
    }
}
```

## Configuration

```go
// Custom base URL and timeout
client, err := paylio.NewClient("sk_live_xxx",
    paylio.WithBaseURL("https://custom.api.com/v1"),
    paylio.WithTimeout(60 * time.Second),
)

// Custom HTTP client
client, err := paylio.NewClient("sk_live_xxx",
    paylio.WithHTTPClient(&http.Client{
        Transport: customTransport,
    }),
)
```

## Error Types

| Error | HTTP Status | Description |
|---|---|---|
| `AuthenticationError` | 401 | Invalid or missing API key |
| `InvalidRequestError` | 400 | Bad request parameters |
| `NotFoundError` | 404 | Resource not found |
| `RateLimitError` | 429 | Rate limit exceeded |
| `APIError` | 5xx | Server error or unexpected status |
| `APIConnectionError` | — | Network failure or timeout |

All error types embed `*PaylioError` and work with `errors.As`.

## License

MIT
