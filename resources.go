package paylio

import (
	"encoding/json"
	"fmt"
)

// Plan represents a subscription plan.
type Plan struct {
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	Interval string  `json:"interval"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Period represents a time period with start and end timestamps.
type Period struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Subscription represents a user's subscription.
type Subscription struct {
	ID                 string  `json:"id"`
	Object             string  `json:"object"`
	Status             string  `json:"status"`
	UserID             string  `json:"user_id"`
	Plan               Plan    `json:"plan"`
	SubscriptionPeriod Period  `json:"subscription_period"`
	CancelAtPeriodEnd  bool    `json:"cancel_at_period_end"`
	CanceledAt         *string `json:"canceled_at"`
	Provider           string  `json:"provider"`
	CreatedAt          string  `json:"created_at"`
}

// SubscriptionCancel represents the result of canceling a subscription.
type SubscriptionCancel struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Success           bool   `json:"success"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
}

// SubscriptionHistoryItem represents a single item in subscription history.
type SubscriptionHistoryItem struct {
	ID                 string  `json:"id"`
	UserID             string  `json:"user_id"`
	PlanSlug           string  `json:"plan_slug"`
	PlanName           string  `json:"plan_name"`
	PlanAmount         float64 `json:"plan_amount"`
	PlanCurrency       string  `json:"plan_currency"`
	PlanInterval       string  `json:"plan_interval"`
	Status             string  `json:"status"`
	CurrentPeriodStart string  `json:"current_period_start"`
	CurrentPeriodEnd   string  `json:"current_period_end"`
	CreatedAt          string  `json:"created_at"`
}

// PaginatedList is a generic paginated response container.
type PaginatedList[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// HasMore returns true if there are additional pages of results.
func (p *PaginatedList[T]) HasMore() bool {
	return p.Page > 0 && p.Page < p.TotalPages
}

// unmarshalTo converts a map[string]any to a typed struct via JSON round-trip.
func unmarshalTo[T any](data map[string]any) (*T, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &result, nil
}
