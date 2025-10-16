package core

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"
)

// TestAVIFEncodeDecodeReal exercises real avifenc/avifdec binaries if available.
func TestAVIFEncodeDecodeReal(t *testing.T) {
	if _, err := exec.LookPath("avifenc"); err != nil {
		t.Skip("avifenc not installed")
	}
	if _, err := exec.LookPath("avifdec"); err != nil {
		t.Skip("avifdec not installed")
	}

	handler := &AvifHandler{}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{10, 20, 30, 255})
	var enc bytes.Buffer
	if err := handler.Encode(&enc, img, 80); err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	decoded, err := handler.Decode(bytes.NewReader(enc.Bytes()))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	// Normalize to RGBA for comparison
	rgba := image.NewRGBA(decoded.Bounds())
	for y := decoded.Bounds().Min.Y; y < decoded.Bounds().Max.Y; y++ {
		for x := decoded.Bounds().Min.X; x < decoded.Bounds().Max.X; x++ {
			rgba.Set(x, y, decoded.At(x, y))
		}
	}
	c := rgba.RGBAAt(0, 0)
	dr := int(c.R) - 10
	dg := int(c.G) - 20
	db := int(c.B) - 30
	if dr*dr+dg*dg+db*db > 25 { // simple squared distance tolerance
		t.Fatalf("pixel mismatch after round trip: %#v (tolerance exceeded)", c)
	}
}

// TestAVIFDecodeInvalid ensures invalid bytes cause decode failure without mocking.
func TestAVIFDecodeInvalid(t *testing.T) {
	if _, err := exec.LookPath("avifdec"); err != nil {
		t.Skip("avifdec not installed")
	}
	handler := &AvifHandler{}
	_, err := handler.Decode(bytes.NewReader([]byte("not-an-avif")))
	if err == nil {
		t.Fatalf("expected error for invalid avif data")
	}
}
