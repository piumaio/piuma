package core_test

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestAdaptiveQualitySkipIfDssimMissing runs adaptive path only if `dssim` present.
func TestAdaptiveQualitySkipIfDssimMissing(t *testing.T) {
	if _, err := exec.LookPath("dssim"); err != nil {
		t.Skip("dssim not installed, skipping adaptive quality test")
	}
	// Build a small RGBA image and run CompressByDSSIM with WebP handler.
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 16), uint8(y * 16), 0, 255})
		}
	}
	var handler core.ImageHandler = &core.WebPHandler{}
	var buf bytes.Buffer
	var imgInterface image.Image = img
	err := core.CompressByDSSIM(&imgInterface, &buf, &handler, 0.005)
	if err != nil {
		t.Fatalf("CompressByDSSIM error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected some bytes written")
	}
}
