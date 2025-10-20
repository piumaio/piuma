package core_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestBuildResponseSuccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "brs*")
	if err != nil {
		t.Fatalf("temp file error: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("hello"))
	tmpFile.Close()
	w := httptest.NewRecorder()
	if err := core.BuildResponse(w, tmpFile.Name(), "text/plain"); err != nil {
		t.Fatalf("build response error: %v", err)
	}
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if body := w.Body.String(); body != "hello" {
		t.Fatalf("expected body 'hello' got %s", body)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("expected content-type text/plain got %s", ct)
	}
}
