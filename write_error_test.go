package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// helper to invoke writeError (in server.go) with controlled response.
func TestWriteErrorBranches(t *testing.T) {
	w := httptest.NewRecorder()
	r := &http.Response{StatusCode: 404, Header: http.Header{}, Request: &http.Request{URL: mustURL("http://domain.test/img.png")}}
	r.Header.Set("Content-Type", "text/plain")
	writeError(w, r, errString("invalid_status_code"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
	assertJSONHas(t, w.Body.Bytes(), "error", "INVALID_STATUS_CODE")

	w = httptest.NewRecorder()
	writeError(w, r, errString("invalid_content_type"))
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 got %d", w.Code)
	}
	assertJSONHas(t, w.Body.Bytes(), "error", "INVALID_CONTENT_TYPE")

	w = httptest.NewRecorder()
	writeError(w, r, errString("invalid_domain"))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
	assertJSONHas(t, w.Body.Bytes(), "error", "INVALID_DOMAIN")

	w = httptest.NewRecorder()
	writeError(w, r, errString("something_else"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
	assertJSONHas(t, w.Body.Bytes(), "error", "SOMETHING_ELSE")
}

// minimal error type implementing Error for passing custom strings
type errString string

func (e errString) Error() string { return string(e) }

func assertJSONHas(t *testing.T, data []byte, key string, expected string) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json parse error: %v", err)
	}
	v, ok := m[key]
	if !ok {
		t.Fatalf("key %s missing", key)
	}
	if v.(string) != expected {
		t.Fatalf("expected %s got %s", expected, v.(string))
	}
}

// local mustURL helper (duplicated minimal version for main package tests)
func mustURL(u string) *url.URL { r, _ := http.NewRequest("GET", u, nil); return r.URL }
