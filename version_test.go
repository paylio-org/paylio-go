package paylio

import "testing"

func TestVersion(t *testing.T) {
	if Version != "0.1.3" {
		t.Errorf("expected version 0.1.2, got %s", Version)
	}
}
