package core_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/piumaio/piuma/core"
)

// TestStartHttpCachePurge ensures expired cache file is removed after purge tick.
func TestStartHttpCachePurge(t *testing.T) {
	// Prepare fake cached file entry
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "fake_cache")
	ioutil.WriteFile(cacheFile, []byte("HTTP/1.1 200 OK\r\nContent-Type: image/png\r\n\r\n"), 0o644)
	core.HttpCacheMutex.Store(cacheFile, time.Now().Unix()-1) // expired

	quit := core.StartHttpCachePurge(1)
	defer func() { quit <- true }()
	// Wait a little over 1s for ticker
	time.Sleep(1100 * time.Millisecond)
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Fatalf("expected cache file to be purged, still exists")
	}
}
