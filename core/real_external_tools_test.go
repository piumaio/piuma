package core_test

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestRealOptiPNG attempts a real PNG optimization if optipng is installed.
func TestRealOptiPNG(t *testing.T) {
	if _, err := exec.LookPath("optipng"); err != nil {
		t.Skip("optipng not installed; skipping real optimization test")
	}
	img := image.NewRGBA(image.Rect(0, 0, 6, 4))
	img.Set(0, 0, color.RGBA{10, 20, 30, 255})
	var buf bytes.Buffer
	handler := &core.PNGHandler{}
	if err := handler.Encode(&buf, img, 0); err != nil {
		t.Fatalf("PNG encode with real optipng failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty optimized PNG output")
	}
	// Basic PNG signature check
	out := buf.Bytes()
	pngSig := []byte{0x89, 'P', 'N', 'G'}
	for i, b := range pngSig {
		if out[i] != b {
			t.Fatalf("PNG signature mismatch")
		}
	}
}

// TestRealJPEGOptim attempts a real JPEG optimization if jpegoptim is installed.
func TestRealJPEGOptim(t *testing.T) {
	if _, err := exec.LookPath("jpegoptim"); err != nil {
		t.Skip("jpegoptim not installed; skipping real optimization test")
	}
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	img.Set(0, 0, color.RGBA{200, 100, 50, 255})
	var buf bytes.Buffer
	handler := &core.JPEGHandler{}
	if err := handler.Encode(&buf, img, 85); err != nil {
		t.Fatalf("JPEG encode with real jpegoptim failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty optimized JPEG output")
	}
	// JPEG files start with 0xFF 0xD8
	out := buf.Bytes()
	if len(out) < 2 || out[0] != 0xFF || out[1] != 0xD8 {
		t.Fatalf("JPEG SOI marker missing")
	}
}
