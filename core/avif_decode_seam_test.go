package core

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"
)

func TestAVIFDecodeSeamFailure(t *testing.T) {
	orig := avifdecCmd
	avifdecCmd = func(args ...string) *exec.Cmd { return exec.Command("false") }
	handler := &AvifHandler{}
	var buf bytes.Buffer
	// encode a png then attempt failing decode path (will return error from avifdec)
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{10, 20, 30, 255})
	// craft fake avif bytes (empty)
	if _, err := handler.Decode(bytes.NewReader(buf.Bytes())); err == nil {
		t.Fatalf("expected decode failure")
	}
	avifdecCmd = orig
}
