package paylio

import (
	"encoding/json"
	"testing"
)

func TestSubscriptionUnmarshal(t *testing.T) {
	raw := `{
		"id": "sub_123",
		"object": "subscription",
		"status": "active",
		"user_id": "user_456",
		"plan": {"slug": "pro", "name": "Pro Plan", "interval": "month", "amount": 999, "currency": "usd"},
		"subscription_period": {"start": "2025-01-01T00:00:00Z", "end": "2025-02-01T00:00:00Z"},
		"cancel_at_period_end": false,
		"canceled_at": null,
		"provider": "stripe",
		"created_at": "2025-01-01T00:00:00Z"
	}`

	var sub Subscription
	if err := json.Unmarshal([]byte(raw), &sub); err != nil {
		t.Fatal(err)
	}

	if sub.ID != "sub_123" {
		t.Errorf("ID = %q", sub.ID)
	}
	if sub.Object != "subscription" {
		t.Errorf("Object = %q", sub.Object)
	}
	if sub.Status != "active" {
		t.Errorf("Status = %q", sub.Status)
	}
	if sub.UserID != "user_456" {
		t.Errorf("UserID = %q", sub.UserID)
	}
	if sub.Plan.Slug != "pro" {
		t.Errorf("Plan.Slug = %q", sub.Plan.Slug)
	}
	if sub.Plan.Name != "Pro Plan" {
		t.Errorf("Plan.Name = %q", sub.Plan.Name)
	}
	if sub.Plan.Interval != "month" {
		t.Errorf("Plan.Interval = %q", sub.Plan.Interval)
	}
	if sub.Plan.Amount != 999 {
		t.Errorf("Plan.Amount = %v", sub.Plan.Amount)
	}
	if sub.Plan.Currency != "usd" {
		t.Errorf("Plan.Currency = %q", sub.Plan.Currency)
	}
	if sub.SubscriptionPeriod.Start != "2025-01-01T00:00:00Z" {
		t.Errorf("Period.Start = %q", sub.SubscriptionPeriod.Start)
	}
	if sub.SubscriptionPeriod.End != "2025-02-01T00:00:00Z" {
		t.Errorf("Period.End = %q", sub.SubscriptionPeriod.End)
	}
	if sub.CancelAtPeriodEnd {
		t.Error("CancelAtPeriodEnd should be false")
	}
	if sub.CanceledAt != nil {
		t.Errorf("CanceledAt = %v", sub.CanceledAt)
	}
	if sub.Provider != "stripe" {
		t.Errorf("Provider = %q", sub.Provider)
	}
	if sub.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q", sub.CreatedAt)
	}
}

func TestSubscriptionCanceledAtNonNull(t *testing.T) {
	raw := `{"id":"sub_1","canceled_at":"2025-03-01T00:00:00Z"}`
	var sub Subscription
	if err := json.Unmarshal([]byte(raw), &sub); err != nil {
		t.Fatal(err)
	}
	if sub.CanceledAt == nil || *sub.CanceledAt != "2025-03-01T00:00:00Z" {
		t.Errorf("CanceledAt = %v", sub.CanceledAt)
	}
}

func TestSubscriptionMarshalRoundTrip(t *testing.T) {
	original := Subscription{
		ID:     "sub_1",
		Status: "active",
		Plan:   Plan{Slug: "basic", Amount: 500, Currency: "usd"},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Subscription
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != original.ID || decoded.Plan.Slug != original.Plan.Slug {
		t.Error("round-trip mismatch")
	}
}

func TestSubscriptionCancelUnmarshal(t *testing.T) {
	raw := `{"id":"sub_1","object":"subscription_cancel","success":true,"cancel_at_period_end":true}`
	var sc SubscriptionCancel
	if err := json.Unmarshal([]byte(raw), &sc); err != nil {
		t.Fatal(err)
	}
	if sc.ID != "sub_1" {
		t.Errorf("ID = %q", sc.ID)
	}
	if !sc.Success {
		t.Error("Success should be true")
	}
	if !sc.CancelAtPeriodEnd {
		t.Error("CancelAtPeriodEnd should be true")
	}
}

func TestSubscriptionHistoryItemUnmarshal(t *testing.T) {
	raw := `{
		"id": "sub_1",
		"user_id": "user_1",
		"plan_slug": "pro",
		"plan_name": "Pro",
		"plan_amount": 999,
		"plan_currency": "usd",
		"plan_interval": "month",
		"status": "active",
		"current_period_start": "2025-01-01T00:00:00Z",
		"current_period_end": "2025-02-01T00:00:00Z",
		"created_at": "2025-01-01T00:00:00Z"
	}`
	var item SubscriptionHistoryItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}
	if item.ID != "sub_1" {
		t.Errorf("ID = %q", item.ID)
	}
	if item.PlanSlug != "pro" {
		t.Errorf("PlanSlug = %q", item.PlanSlug)
	}
	if item.PlanAmount != 999 {
		t.Errorf("PlanAmount = %v", item.PlanAmount)
	}
	if item.Status != "active" {
		t.Errorf("Status = %q", item.Status)
	}
}

func TestPaginatedListHasMore(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		totalPages int
		want       bool
	}{
		{"more pages", 1, 3, true},
		{"last page", 3, 3, false},
		{"zero page", 0, 0, false},
		{"single page", 1, 1, false},
		{"page 2 of 5", 2, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := PaginatedList[SubscriptionHistoryItem]{
				Page:       tt.page,
				TotalPages: tt.totalPages,
			}
			if got := pl.HasMore(); got != tt.want {
				t.Errorf("HasMore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaginatedListUnmarshal(t *testing.T) {
	raw := `{
		"items": [{"id": "sub_1", "status": "active"}, {"id": "sub_2", "status": "canceled"}],
		"total": 2,
		"page": 1,
		"page_size": 20,
		"total_pages": 1
	}`
	var pl PaginatedList[SubscriptionHistoryItem]
	if err := json.Unmarshal([]byte(raw), &pl); err != nil {
		t.Fatal(err)
	}
	if pl.Total != 2 {
		t.Errorf("Total = %d", pl.Total)
	}
	if len(pl.Items) != 2 {
		t.Errorf("Items len = %d", len(pl.Items))
	}
	if pl.Items[0].ID != "sub_1" {
		t.Errorf("Items[0].ID = %q", pl.Items[0].ID)
	}
	if pl.Items[1].Status != "canceled" {
		t.Errorf("Items[1].Status = %q", pl.Items[1].Status)
	}
}

func TestUnmarshalToSuccess(t *testing.T) {
	data := map[string]any{"id": "sub_1", "status": "active"}
	result, err := unmarshalTo[Subscription](data)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "sub_1" {
		t.Errorf("ID = %q", result.ID)
	}
}
