package core

import (
	"net/http"
	"net/url"
	"testing"
)

func TestWorkerManagerLifecycle(t *testing.T) {
	wm := NewWorkerManager()
	wm.Run()
	if wm.closed {
		t.Fatalf("expected manager open")
	}
	wm.Close()
	if !wm.closed {
		t.Fatalf("expected manager closed after Close")
	}
	// Dispatch after close returns nil
	ch := wm.Dispatch(&http.Response{}, &ImageParameters{}, &Options{})
	if ch != nil {
		t.Fatalf("expected nil channel after close")
	}
}

func TestAsyncOptimizeTimeoutNoWorkers(t *testing.T) {
	GlobalWorkerManager = NewWorkerManager()
	u, _ := url.Parse("http://example.com/img.png")
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: u}}
	resp.Header.Set("Content-Type", "image/png")
	params := &ImageParameters{Quality: 80}
	_, _, err := asyncOptimize(resp, params, &Options{Timeout: 1})
	if err == nil || err.Error() != "timed out" {
		t.Fatalf("expected timed out got %v", err)
	}
}
