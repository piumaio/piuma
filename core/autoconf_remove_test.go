package core_test

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/elnormous/contenttype"
	"github.com/piumaio/piuma/core"
)

// TestRemoveImageHandlerFromAutoConf ensures a format is removed from the autoconf list.
func TestRemoveImageHandlerFromAutoConf(t *testing.T) {
	tmpDir := t.TempDir()
	autoConfPath := tmpDir + "/autoconf.gob"

	upstreamURL, _ := url.Parse("http://example.com/img.png")
	upstream := &http.Response{Header: http.Header{}, Request: &http.Request{URL: upstreamURL}}
	upstream.Header.Set("Content-Type", "image/png")

	req := &http.Request{Header: http.Header{}}
	req.Header.Set("Accept", "image/png,image/webp,image/avif")

	// First call creates the file and list.
	_, err := core.AutoImageHandler(req, upstream, autoConfPath, []string{})
	if err != nil {
		t.Fatalf("auto image handler error: %v", err)
	}

	// Remove webp
	if err := core.RemoveImageHandlerFromAutoConf(autoConfPath, "image/webp"); err != nil {
		t.Fatalf("remove error: %v", err)
	}

	// Read back and ensure webp gone
	f, err := os.Open(autoConfPath)
	if err != nil {
		t.Fatalf("open autoconf: %v", err)
	}
	defer f.Close()
	var mts []contenttype.MediaType
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&mts); err != nil {
		t.Fatalf("decode gob: %v", err)
	}
	for _, mt := range mts {
		if mt.String() == "image/webp" {
			t.Fatalf("webp still present after removal")
		}
	}
}
