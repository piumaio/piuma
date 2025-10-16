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

func TestOptimizePenalizeFormat(t *testing.T) {
	tmpDir := t.TempDir()
	// Create tiny original temp file smaller than encoded output
	tempFile := filepath.Join(tmpDir, "in.webp")
	os.WriteFile(tempFile, []byte("x"), 0o644)

	// Build response + params
	resp := &http.Response{Header: http.Header{}, Request: &http.Request{URL: mustURL("http://example.com/x.webp")}}
	resp.Header.Set("Content-Type", "image/webp")
	params, _ := core.Parser("2_2_90:webp")

	// Create image file expected by Decode (replace temp file with valid webp bytes)
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{200, 0, 0, 255})
	f, _ := os.Create(tempFile)
	webp.Encode(f, img, &webp.Options{Quality: 90})
	f.Close()

	outPath := filepath.Join(tmpDir, "out.webp")
	_, mime, err := core.Optimize(resp, &params, &core.Options{PathTemp: tempFile, PathMedia: outPath})
	if err != nil {
		t.Fatalf("optimize error: %v", err)
	}
	if mime != "image/webp" {
		t.Fatalf("expected mime image/webp got %s", mime)
	}
}

// helper moved to test_helpers_test.go
