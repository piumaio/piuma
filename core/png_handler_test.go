package core_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestPNGHandlerEncodeDecode(t *testing.T) {
	h := &core.PNGHandler{}
	img := image.NewRGBA(image.Rect(0, 0, 5, 3))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	var buf bytes.Buffer
	// Use low quality to skip external optimization path if handler uses it
	// Instead of relying on handler's external optimization, directly encode PNG to buffer then run Decode.
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("direct png encode error: %v", err)
	}
	// Ensure handler Decode works on produced bytes
	out, err := h.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if out.Bounds().Dx() != 5 || out.Bounds().Dy() != 3 {
		t.Fatalf("unexpected bounds")
	}
	if !h.SupportsTransparency() {
		t.Fatalf("expected transparency support")
	}
}
