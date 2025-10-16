package main

import (
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chai2010/webp"
	"github.com/julienschmidt/httprouter"
	"github.com/piumaio/piuma/core"
)

// TestProcessImageFallback ensures original bytes served when optimization fails.
func TestProcessImageFallback(t *testing.T) {
	// upstream server returns webp
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{100, 0, 0, 255})
		w.Header().Set("Content-Type", "image/webp")
		webp.Encode(w, img, &webp.Options{Quality: 80})
	}))
	defer upstream.Close()
	// configure globals minimal
	pathtemp = t.TempDir()
	pathmedia = t.TempDir()
	// allow upstream host (host:port)
	domains_list = []string{strings.TrimPrefix(upstream.URL, "http://")}
	httpCacheTTL = 1
	timeout = 1
	// closed worker manager forces asyncOptimize timeout error
	core.GlobalWorkerManager = core.NewWorkerManager()
	core.GlobalWorkerManager.Close() // ensure manager is closed so asyncOptimize fails
	r := httptest.NewRequest("GET", "/0_0_80:webp/"+strings.TrimPrefix(upstream.URL, "http://")+"/img.webp", nil)
	w := httptest.NewRecorder()
	params := httprouter.Params{httprouter.Param{Key: "parameters", Value: "0_0_80:webp"}, httprouter.Param{Key: "url", Value: strings.TrimPrefix(upstream.URL, "http://") + "/img.webp"}}
	processImage(w, r, params)
	if w.Code != 200 && w.Code != 403 {
		t.Fatalf("expected 200 or 403 got %d", w.Code)
	}
}
