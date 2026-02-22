package paylio

import "testing"

func TestVersion(t *testing.T) {
	if Version != "0.1.1" {
		t.Errorf("expected version 0.1.1, got %s", Version)
	}
}
