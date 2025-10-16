package core_test

import (
	"net/http"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestGetUrlString(t *testing.T) {
	params, err := core.Parser("100_50_75a:webp")
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	s := params.GetUrlString()
	if s != "100_50_75a:webp" {
		t.Fatalf("expected canonical string got %s", s)
	}
	// hash influenced by convert & adaptive flags
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: mustURL("http://example.com/x.png")}}
	resp.Header.Set("Content-Type", "image/png")
	h := params.GenerateHash(resp)
	if h == "" {
		t.Fatalf("expected non-empty hash")
	}
}
