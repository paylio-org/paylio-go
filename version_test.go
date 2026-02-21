package paylio

import "testing"

func TestVersion(t *testing.T) {
	if Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", Version)
	}
}
