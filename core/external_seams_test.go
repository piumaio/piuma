package core

import (
	"bytes"
	"image"
	"image/color"
	"os/exec"
	"testing"
)

// TestJPEGEncodeSeams simulates success and failure without real jpegoptim.
func TestJPEGEncodeSeams(t *testing.T) {
	// build sample image
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{200, 0, 0, 255})
	var buf bytes.Buffer
	// success: override to 'true'
	orig := jpegoptimCmd
	jpegoptimCmd = func(args ...string) *exec.Cmd { return exec.Command("true") }
	handler := &JPEGHandler{}
	if err := handler.Encode(&buf, img, 80); err != nil {
		t.Fatalf("expected success encode got %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected some jpeg bytes")
	}
	// failure: override to 'false'
	jpegoptimCmd = func(args ...string) *exec.Cmd { return exec.Command("false") }
	buf.Reset()
	if err := handler.Encode(&buf, img, 80); err == nil {
		t.Fatalf("expected failure from false command")
	}
	// restore
	jpegoptimCmd = orig
}

// TestAVIFEncodeSeam simulates missing avifenc tool.
func TestAVIFEncodeSeam(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	handler := &AvifHandler{}
	orig := avifencCmd
	avifencCmd = func(args ...string) *exec.Cmd { return exec.Command("false") }
	var buf bytes.Buffer
	if err := handler.Encode(&buf, img, 50); err == nil {
		t.Fatalf("expected avif encode failure when command false")
	}
	avifencCmd = orig
}

// TestDSSIMSeamSuccess crafts dssim output to drive adaptive loop.
func TestDSSIMSeamSuccess(t *testing.T) {
	orig := dssimCmd
	dssimCmd = func(args ...string) *exec.Cmd { return exec.Command("bash", "-c", "echo '0.0005\txyz'") }
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var handler ImageHandler = &WebPHandler{}
	var out bytes.Buffer
	var imgI image.Image = img
	if err := CompressByDSSIM(&imgI, &out, &handler, 0.001); err != nil {
		t.Fatalf("expected dssim adaptive success got %v", err)
	}
	if out.Len() == 0 {
		t.Fatalf("expected output bytes")
	}
	dssimCmd = orig
}

// TestDSSIMSeamFailure ensures failure path returns error.
func TestDSSIMSeamFailure(t *testing.T) {
	orig := dssimCmd
	dssimCmd = func(args ...string) *exec.Cmd { return exec.Command("false") }
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var handler ImageHandler = &WebPHandler{}
	var out bytes.Buffer
	var imgI image.Image = img
	if err := CompressByDSSIM(&imgI, &out, &handler, 0.001); err == nil {
		t.Fatalf("expected error from dssim false cmd")
	}
	dssimCmd = orig
}
