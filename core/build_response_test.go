package core_test

import (
	"net/http/httptest"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestBuildResponseError ensures error when optimized file missing.
func TestBuildResponseError(t *testing.T) {
	w := httptest.NewRecorder()
	err := core.BuildResponse(w, "nonexistent_file.xyz", "image/webp")
	if err == nil || err.Error() != "error reading from optimized file" {
		t.Fatalf("expected file read error got %v", err)
	}
}
