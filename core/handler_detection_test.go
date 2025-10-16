package core_test

import (
	"bytes"
	"testing"

	"github.com/piumaio/piuma/core"
)

// TestNewImageHandlerByBytesAVIF crafts a minimal byte slice that should be detected as AVIF.
func TestNewImageHandlerByBytesAVIF(t *testing.T) {
	buf := make([]byte, 512)
	copy(buf[8:12], []byte("avif")) // trigger special-case in detection
	handler, err := core.NewImageHandlerByBytes(bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handler.ImageExtension() != "avif" {
		t.Fatalf("expected avif extension, got %s", handler.ImageExtension())
	}
}
