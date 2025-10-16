package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestParseConfigAndInitialize(t *testing.T) {
    tempDir := t.TempDir()
    args := []string{"-port", "9090", "-mediapath", filepath.Join(tempDir, "media"), "-timeout", "5", "-httpCacheTTL", "10", "-httpCachePurgeInterval", "20", "-workers", "2", "-domains", "example.com,example.org"}
    cfg, err := parseConfig(args)
    if err != nil { t.Fatalf("parseConfig error: %v", err) }
    if cfg.Port != "9090" || cfg.Timeout != 5 || cfg.HTTPCacheTTL != 10 || cfg.Workers != 2 { t.Fatalf("config values not parsed correctly: %+v", cfg) }
    router, shutdown, err := Initialize(cfg)
    if err != nil { t.Fatalf("initialize error: %v", err) }
    defer shutdown()
    // verify globals
    if pathmedia != cfg.MediaPath { t.Fatalf("expected pathmedia=%s got %s", cfg.MediaPath, pathmedia) }
    if pathtemp != filepath.Join(cfg.MediaPath, "temp") { t.Fatalf("pathtemp incorrect: %s", pathtemp) }
    if len(domains_list) != 2 { t.Fatalf("expected 2 domains parsed got %d", len(domains_list)) }
    // directories exist
    if _, err := os.Stat(pathtemp); err != nil { t.Fatalf("temp path missing: %v", err) }
    // basic route presence check (router is from httprouter; we rely on internal map length >0)
    if router == nil { t.Fatalf("router was nil") }
}

func TestRunInvalidFlag(t *testing.T) {
    err := run([]string{"-notaflag"})
    if err == nil { t.Fatalf("expected error for invalid flag") }
}
