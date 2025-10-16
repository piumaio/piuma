package core_test

import (
	"net/http"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestIsImage(t *testing.T) {
	rImg := &http.Response{Header: http.Header{}}
	rImg.Header.Set("Content-Type", "image/png")
	if !core.IsImage(rImg) {
		t.Fatalf("expected true for image content-type")
	}
	rTxt := &http.Response{Header: http.Header{}}
	rTxt.Header.Set("Content-Type", "text/plain")
	if core.IsImage(rTxt) {
		t.Fatalf("expected false for non-image content-type")
	}
}
