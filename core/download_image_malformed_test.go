package core_test

import (
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestDownloadImageMalformedURL(t *testing.T) {
	resp, err := core.DownloadImage("bad", 1, []string{"example.com"})
	if err == nil || err.Error() != "invalid_domain" {
		t.Fatalf("expected invalid_domain for malformed url got %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 status got %d", resp.StatusCode)
	}
}
