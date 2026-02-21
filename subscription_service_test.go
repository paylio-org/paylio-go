package paylio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestService(handler http.HandlerFunc) (*SubscriptionService, *httptest.Server) {
	srv := httptest.NewServer(handler)
	hc := newHTTPClient("sk_test", srv.URL, 10*time.Second, srv.Client())
	return newSubscriptionService(hc), srv
}

func TestRetrieveReturnsSubscription(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q", r.Method)
		}
		if r.URL.Path != "/subscription/user_123" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"sub_1","status":"active","user_id":"user_123","plan":{"slug":"pro"}}`))
	})
	defer srv.Close()

	sub, err := svc.Retrieve(context.Background(), "user_123")
	if err != nil {
		t.Fatal(err)
	}
	if sub.ID != "sub_1" {
		t.Errorf("ID = %q", sub.ID)
	}
	if sub.Status != "active" {
		t.Errorf("Status = %q", sub.Status)
	}
	if sub.Plan.Slug != "pro" {
		t.Errorf("Plan.Slug = %q", sub.Plan.Slug)
	}
}

func TestRetrieveEmptyUserIDReturnsError(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := svc.Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
	if err.Error() != "userID is required" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestRetrieveWhitespaceUserIDReturnsError(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := svc.Retrieve(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for whitespace userID")
	}
}

func TestListReturnsPaginatedList(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q", r.Method)
		}
		if r.URL.Path != "/users/user_1/subscriptions" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "5" {
			t.Errorf("page_size = %q", r.URL.Query().Get("page_size"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"items":[{"id":"sub_1","status":"active"},{"id":"sub_2","status":"canceled"}],"total":10,"page":2,"page_size":5,"total_pages":2}`))
	})
	defer srv.Close()

	list, err := svc.List(context.Background(), "user_1", &ListOptions{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 2 {
		t.Errorf("Items len = %d", len(list.Items))
	}
	if list.Items[0].ID != "sub_1" {
		t.Errorf("Items[0].ID = %q", list.Items[0].ID)
	}
	if list.Total != 10 {
		t.Errorf("Total = %d", list.Total)
	}
	if list.HasMore() {
		t.Error("HasMore should be false (page 2 of 2)")
	}
}

func TestListDefaultPagination(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("page = %q", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "20" {
			t.Errorf("page_size = %q", r.URL.Query().Get("page_size"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"items":[],"total":0,"page":1,"page_size":20,"total_pages":0}`))
	})
	defer srv.Close()

	_, err := svc.List(context.Background(), "user_1", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListEmptyUserIDReturnsError(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := svc.List(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
}

func TestListHasMoreTrue(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"items":[],"total":50,"page":1,"page_size":20,"total_pages":3}`))
	})
	defer srv.Close()

	list, err := svc.List(context.Background(), "user_1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !list.HasMore() {
		t.Error("HasMore should be true (page 1 of 3)")
	}
}

func TestCancelReturnsSubscriptionCancel(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q", r.Method)
		}
		if r.URL.Path != "/subscription/sub_uuid/cancel" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatal(err)
		}
		// Default: cancelNow=false -> cancel_at_period_end=true
		if parsed["cancel_at_period_end"] != true {
			t.Errorf("cancel_at_period_end = %v", parsed["cancel_at_period_end"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"sub_uuid","object":"subscription_cancel","success":true,"cancel_at_period_end":true}`))
	})
	defer srv.Close()

	result, err := svc.Cancel(context.Background(), "sub_uuid", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if !result.CancelAtPeriodEnd {
		t.Error("CancelAtPeriodEnd should be true")
	}
}

func TestCancelNowSendsFalse(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatal(err)
		}
		// cancelNow=true -> cancel_at_period_end=false
		if parsed["cancel_at_period_end"] != false {
			t.Errorf("cancel_at_period_end = %v", parsed["cancel_at_period_end"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"sub_uuid","success":true,"cancel_at_period_end":false}`))
	})
	defer srv.Close()

	_, err := svc.Cancel(context.Background(), "sub_uuid", &CancelOptions{CancelNow: true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelEmptySubscriptionIDReturnsError(t *testing.T) {
	svc, srv := newTestService(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := svc.Cancel(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty subscriptionID")
	}
}
