package core_test

import (
	"image"
	"image/color"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/chai2010/webp"
	"github.com/piumaio/piuma/core"
)

func TestOptimizeWebP(t *testing.T) {
	// Create a simple 10x10 RGBA image and encode as webp (library based)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 25), uint8(y * 25), 0, 255})
		}
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "in.webp")
	f, _ := os.Create(tmpFile)
	webp.Encode(f, img, &webp.Options{Lossless: false, Quality: 90})
	f.Close()

	respURL, _ := url.Parse("http://example.com/img.webp")
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: respURL}}
	resp.Header.Set("Content-Type", "image/webp")

	params, _ := core.Parser("5_5_80")
	outPath := filepath.Join(tmpDir, "out.webp")
	options := &core.Options{PathTemp: tmpFile, PathMedia: outPath}

	path, mime, err := core.Optimize(resp, &params, options)
	if err != nil {
		t.Fatalf("optimize error: %v", err)
	}
	if path != outPath {
		t.Fatalf("expected output path %s got %s", outPath, path)
	}
	if mime != "image/webp" {
		t.Fatalf("expected mime image/webp got %s", mime)
	}

	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatalf("output file not created")
	}
}
