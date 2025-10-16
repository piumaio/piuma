package core_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestDownloadImageCacheReuse(t *testing.T) {
	body := []byte("123456789012345")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(body)
	}))
	full := srv.URL + "/a.png"
	parsed, _ := url.Parse(full)
	resp1, err := core.DownloadImage(full, 10, []string{parsed.Host})
	if err != nil {
		t.Fatalf("first download error: %v", err)
	}
	// Close server to force cache path
	srv.Close()
	resp2, err := core.DownloadImage(full, 10, []string{parsed.Host})
	if err != nil {
		t.Fatalf("second download (cache) error: %v", err)
	}
	if resp2.StatusCode != 200 {
		t.Fatalf("expected cached 200 got %d", resp2.StatusCode)
	}
	resp1.Body.Close()
	resp2.Body.Close()
}
