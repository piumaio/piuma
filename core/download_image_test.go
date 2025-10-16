package core_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestDownloadImageDomainValidation checks allow-list enforcement.
func TestDownloadImageDomainValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
		w.Write([]byte("123"))
	}))
	defer srv.Close()

	_, err := core.DownloadImage(srv.URL+"/x.png", 0, []string{"example.com"})
	if err == nil || err.Error() != "invalid_domain" {
		t.Fatalf("expected invalid_domain error")
	}

	// Use exact host:port from server URL for allowed domain
	allowedDomain := strings.Split(srv.URL, "/")[2]
	resp, err := core.DownloadImage(srv.URL+"/x.png", 0, []string{allowedDomain})
	if err != nil {
		t.Fatalf("unexpected error with allowed domain: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 remote status")
	}
}
