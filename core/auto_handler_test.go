package core_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestAutoImageHandlerPreferred ensures preferred list restricts negotiation.
func TestAutoImageHandlerPreferred(t *testing.T) {
	upstreamURL, _ := url.Parse("http://example.com/img.png")
	upstream := &http.Response{Header: http.Header{}, Request: &http.Request{URL: upstreamURL}}
	upstream.Header.Set("Content-Type", "image/png")

	req := httptest.NewRequest("GET", "http://localhost/image", nil)
	req.Header.Set("Accept", "image/webp, image/avif")

	handler, err := core.AutoImageHandler(req, upstream, t.TempDir()+"/autoconf", []string{"avif"})
	if err != nil {
		t.Fatalf("auto handler error: %v", err)
	}
	if handler.ImageExtension() != "avif" {
		t.Fatalf("expected avif chosen, got %s", handler.ImageExtension())
	}
}
