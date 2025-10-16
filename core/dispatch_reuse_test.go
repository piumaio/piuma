package core_test

import (
	"image"
	"image/color"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/chai2010/webp"
	"github.com/piumaio/piuma/core"
)

func TestDispatchReuseExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tempPath := filepath.Join(tmpDir, "temp")
	mediaPath := filepath.Join(tmpDir, "media")
	os.MkdirAll(tempPath, 0o755)
	os.MkdirAll(mediaPath, 0o755)

	// Create existing optimized file
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	existingPath := filepath.Join(mediaPath, "existing")
	f, _ := os.Create(existingPath)
	webp.Encode(f, img, &webp.Options{Quality: 80})
	f.Close()

	// Fake response to derive hash matching existingPath name by forcing GenerateHash to that name (simplify by renaming file to hash later)
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: mustURL("http://example.com/a.webp")}}
	resp.Header.Set("Content-Type", "image/webp")
	params, _ := core.Parser("0_0_80")
	hash := params.GenerateHash(resp)
	targetPath := filepath.Join(mediaPath, hash)
	os.Rename(existingPath, targetPath)

	path, mime, err := core.Dispatch(&http.Request{}, resp, &params, &core.Options{PathTemp: tempPath, PathMedia: mediaPath, Timeout: 0})
	if err != nil {
		t.Fatalf("dispatch error: %v", err)
	}
	if path != targetPath {
		t.Fatalf("expected reuse path %s got %s", targetPath, path)
	}
	if mime != "image/webp" {
		t.Fatalf("expected mime image/webp got %s", mime)
	}
}

// helper moved to test_helpers_test.go
