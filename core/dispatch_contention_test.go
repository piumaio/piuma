package core_test

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/chai2010/webp"
	"github.com/piumaio/piuma/core"
)

func TestDispatchContention(t *testing.T) {
	tmp := t.TempDir()
	temp := filepath.Join(tmp, "temp")
	media := filepath.Join(tmp, "media")
	os.MkdirAll(temp, 0o755)
	os.MkdirAll(media, 0o755)

	// Create shared response body
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	var buf bytes.Buffer
	webp.Encode(&buf, img, &webp.Options{Quality: 80})
	resp := &http.Response{Header: http.Header{}, Body: ioNopCloser(&buf), Request: &http.Request{URL: mustURL("http://example.com/a.webp")}}
	resp.Header.Set("Content-Type", "image/webp")
	params, _ := core.Parser("0_0_80")
	opts := &core.Options{PathTemp: temp, PathMedia: media, Timeout: 10}

	// Inject latency so both goroutines race on the same FileMutex entry.
	originalOptimize := core.OptimizeFunc
	core.OptimizeFunc = func(r *http.Response, ip *core.ImageParameters, o *core.Options) (string, string, error) {
		time.Sleep(20 * time.Millisecond)
		return originalOptimize(r, ip, o)
	}
	defer func() { core.OptimizeFunc = originalOptimize }()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := core.Dispatch(&http.Request{}, resp, &params, opts)
			if err != nil && err.Error() == "still elaborating" {
				mu.Lock()
				firstErr = err
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firstErr == nil {
		t.Fatalf("expected one call to return Still elaborating")
	}
}

// simple io.ReadCloser wrapper
type rc struct{ *bytes.Reader }

func ioNopCloser(b *bytes.Buffer) io.ReadCloser { return io.NopCloser(bytes.NewReader(b.Bytes())) }
