package core_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestGenerateHash ensures hash changes when relevant response header changes.
func TestGenerateHash(t *testing.T) {
	params, err := core.Parser("100_0_80:webp")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	baseURL, _ := url.Parse("http://example.com/image.png")

	resp1 := &http.Response{Header: http.Header{}, Request: &http.Request{URL: baseURL}}
	resp1.Header.Set("Content-Type", "image/png")
	resp1.Header.Set("Content-Length", "10")
	resp1.Header.Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")

	resp2 := &http.Response{Header: http.Header{}, Request: &http.Request{URL: baseURL}}
	resp2.Header.Set("Content-Type", "image/png")
	resp2.Header.Set("Content-Length", "11") // changed size
	resp2.Header.Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")

	hash1 := params.GenerateHash(resp1)
	hash2 := params.GenerateHash(resp2)

	if hash1 == hash2 {
		t.Fatalf("expected different hashes when Content-Length changes")
	}
}
