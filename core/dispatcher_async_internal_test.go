package core

import (
	"net/http"
	"net/url"
	"testing"
)

// TestAsyncOptimizeTimeout ensures timed out error produced when manager closed.
func TestAsyncOptimizeTimeout(t *testing.T) {
	GlobalWorkerManager = &WorkerManager{closed: true}
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: mustURLInternal("http://example.com/img.webp")}}
	resp.Header.Set("Content-Type", "image/webp")
	params := ImageParameters{Width: 0, Height: 0, Quality: 80}
	_, _, err := asyncOptimize(resp, &params, &Options{Timeout: 10})
	if err == nil || err.Error() != "timed out" {
		t.Fatalf("expected timed out error got %v", err)
	}
}

func mustURLInternal(u string) *url.URL { r, _ := http.NewRequest("GET", u, nil); return r.URL }
