package core_test

import (
	"net/http"
	"net/url"
)

// mustURL creates a *url.URL without failing the test, used across multiple test files.
func mustURL(u string) *url.URL { r, _ := http.NewRequest("GET", u, nil); return r.URL }
