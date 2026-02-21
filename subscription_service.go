package paylio

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ListOptions configures pagination for subscription list requests.
type ListOptions struct {
	Page     int
	PageSize int
}

// CancelOptions configures subscription cancellation behavior.
type CancelOptions struct {
	CancelNow bool
}

// SubscriptionService provides methods for interacting with subscriptions.
type SubscriptionService struct {
	http *httpClient
}

func newSubscriptionService(hc *httpClient) *SubscriptionService {
	return &SubscriptionService{http: hc}
}

// Retrieve fetches the current subscription for a user.
func (s *SubscriptionService) Retrieve(ctx context.Context, userID string) (*Subscription, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID is required")
	}
	data, err := s.http.request(ctx, "GET", fmt.Sprintf("/subscription/%s", userID), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalTo[Subscription](data)
}

// List fetches paginated subscription history for a user.
func (s *SubscriptionService) List(ctx context.Context, userID string, opts *ListOptions) (*PaginatedList[SubscriptionHistoryItem], error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID is required")
	}
	page := 1
	pageSize := 20
	if opts != nil {
		if opts.Page > 0 {
			page = opts.Page
		}
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
	}
	params := map[string]string{
		"page":      strconv.Itoa(page),
		"page_size": strconv.Itoa(pageSize),
	}
	data, err := s.http.request(ctx, "GET", fmt.Sprintf("/users/%s/subscriptions", userID), &requestOptions{Params: params})
	if err != nil {
		return nil, err
	}
	return unmarshalTo[PaginatedList[SubscriptionHistoryItem]](data)
}

// Cancel cancels a subscription. By default cancels at end of billing period.
// Set CancelOptions.CancelNow to true for immediate cancellation.
func (s *SubscriptionService) Cancel(ctx context.Context, subscriptionID string, opts *CancelOptions) (*SubscriptionCancel, error) {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil, errors.New("subscriptionID is required")
	}
	cancelNow := false
	if opts != nil {
		cancelNow = opts.CancelNow
	}
	body := map[string]any{"cancel_at_period_end": !cancelNow}
	data, err := s.http.request(ctx, "POST", fmt.Sprintf("/subscription/%s/cancel", subscriptionID), &requestOptions{JSONBody: body})
	if err != nil {
		return nil, err
	}
	return unmarshalTo[SubscriptionCancel](data)
}
