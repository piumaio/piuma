package core_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestDownloadImageErrors(t *testing.T) {
	// invalid status code
	srv404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) }))
	defer srv404.Close()
	_, err := core.DownloadImage(srv404.URL+"/x.png", 1, []string{hostPort(srv404.URL)})
	if err == nil || err.Error() != "invalid_status_code" {
		t.Fatalf("expected invalid_status_code, got %v", err)
	}

	// invalid content type
	srvTxt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hi"))
	}))
	defer srvTxt.Close()
	_, err = core.DownloadImage(srvTxt.URL+"/x.txt", 1, []string{hostPort(srvTxt.URL)})
	if err == nil || err.Error() != "invalid_content_type" {
		t.Fatalf("expected invalid_content_type got %v", err)
	}
}

func hostPort(u string) string { return u[len("http://"):] }
