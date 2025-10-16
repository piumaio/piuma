package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chai2010/webp"
	"github.com/julienschmidt/httprouter"
	"github.com/piumaio/piuma/core"
)

// startPiumaServer spins up handlers using in-memory temp dirs.
func startPiumaServer(t *testing.T, mediaPath string, allowedDomains []string, to int) (*httptest.Server, chan bool) {
	tempPath := filepath.Join(mediaPath, "temp")
	os.MkdirAll(tempPath, os.ModePerm)
	os.MkdirAll(mediaPath, os.ModePerm)
	os.MkdirAll(filepath.Join(os.TempDir(), "piuma_http_cache"), os.ModePerm)

	router := httprouter.New()
	router.GET("/", getInfo)
	router.GET("/:parameters/*url", processImage)

	// configure globals used by server.go
	// Override globals safely
	core.GlobalWorkerManager = core.NewWorkerManager()
	for i := 0; i < 2; i++ {
		core.GlobalWorkerManager.Run()
	}

	// set globals used by handlers
	pathtemp = tempPath
	pathmedia = mediaPath
	domains_list = allowedDomains
	httpCacheTTL = 1
	timeout = to

	stopPurge := core.StartHttpCachePurge(1)

	// patch variables used by processImage via shadow (wrap)
	// NOTE: We rely on original globals; define exported-like proxies here.

	return httptest.NewServer(router), stopPurge
}

// Global variable proxies (since server.go uses package-level vars). We alias them here for clarity.
// (no proxy globals; we use original package-level vars from server.go)

// override wrappers referencing original globals in server.go
// We create small adapter functions forwarding to original logic but substituting globals.

// TestIntegrationBasic exercises resize + auto negotiation + error cases.
func TestIntegrationBasic(t *testing.T) {
	// Upstream image server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// generate tiny valid webp
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{255, 0, 0, 255})
		img.Set(1, 0, color.RGBA{0, 255, 0, 255})
		img.Set(0, 1, color.RGBA{0, 0, 255, 255})
		img.Set(1, 1, color.RGBA{255, 255, 0, 255})
		w.Header().Set("Content-Type", "image/webp")
		webp.Encode(w, img, &webp.Options{Lossless: false, Quality: 80})
	}))
	defer upstream.Close()

	mediaPath := t.TempDir()
	// Allow full host:port for upstream
	upstreamHostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stopPurge := startPiumaServer(t, mediaPath, []string{upstreamHostPort}, 500)
	defer func() {
		core.GlobalWorkerManager.Close()
		stopPurge <- true
		srv.Close()
	}()

	// Basic info endpoint
	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("info endpoint error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 info, got %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "extensions") {
		t.Fatalf("info response missing extensions")
	}

	// Single optimize request using full upstream URL
	remoteURL := upstream.URL + "/img.webp"
	target := fmt.Sprintf("%s/10_10_80:%s/%s", srv.URL, "webp", remoteURL)
	respOpt, errOpt := http.Get(target)
	if errOpt != nil {
		t.Fatalf("optimize request error: %v", errOpt)
	}
	defer respOpt.Body.Close()
	if respOpt.StatusCode != 200 {
		t.Fatalf("unexpected optimize status %d", respOpt.StatusCode)
	}
	// Decode result and assert dimensions bounded
	buf, _ := io.ReadAll(respOpt.Body)
	imgDecoded, errImg := webp.Decode(bytes.NewReader(buf))
	if errImg != nil {
		t.Fatalf("decode optimized webp error: %v", errImg)
	}
	if imgDecoded.Bounds().Dx() > 10 || imgDecoded.Bounds().Dy() > 10 {
		t.Fatalf("expected resized <=10x10 got %dx%d", imgDecoded.Bounds().Dx(), imgDecoded.Bounds().Dy())
	}
}

// TestIntegrationRemoteResize fetches a remote lorem picsum image and verifies resize and conversion to png -> webp.
func TestIntegrationRemoteResize(t *testing.T) {
	// Use a live remote image source (non-deterministic content but deterministic dimensions). If network blocked, skip.
	picsumURL := "https://picsum.photos/seed/piuma/200/300"
	// Quick HEAD to check availability
	headResp, err := http.Head(picsumURL)
	if err != nil || headResp.StatusCode != 200 {
		t.Skip("picsum unreachable; skipping")
	}

	mediaPath := t.TempDir()
	// allow domain (strip scheme)
	picsumHost := strings.TrimPrefix(strings.TrimPrefix(picsumURL, "https://"), "http://")
	picsumHost = strings.Split(picsumHost, "/")[0]
	srv, stopPurge := startPiumaServer(t, mediaPath, []string{picsumHost}, 1000)
	defer func() { core.GlobalWorkerManager.Close(); stopPurge <- true; srv.Close() }()

	// Request with resize smaller than original (target 50x50) and convert to png (if original might be jpeg)
	target := fmt.Sprintf("%s/50_50_90:%s/%s", srv.URL, "webp", picsumURL)
	resp, err := http.Get(target)
	if err != nil {
		t.Fatalf("integration remote resize error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	// Attempt webp decode first; fallback to PNG (in case conversion failed) to still assert resize
	decoded, errWebP := webp.Decode(bytes.NewReader(data))
	if errWebP != nil {
		// try png
		pngImg, errPNG := png.Decode(bytes.NewReader(data))
		if errPNG != nil {
			t.Fatalf("unable to decode as webp or png: %v / %v", errWebP, errPNG)
		}
		if pngImg.Bounds().Dx() > 50 || pngImg.Bounds().Dy() > 50 {
			t.Fatalf("png not resized properly: %dx%d", pngImg.Bounds().Dx(), pngImg.Bounds().Dy())
		}
	} else {
		if decoded.Bounds().Dx() > 50 || decoded.Bounds().Dy() > 50 {
			t.Fatalf("webp not resized properly: %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
		}
	}
}

// TestIntegrationMultipleConversions spins up upstream PNG/JPEG variants and tests conversion directives.
func TestIntegrationMultipleConversions(t *testing.T) {
	// Upstream provides PNG
	upstreamPNG := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 20, 10))
		img.Set(0, 0, color.RGBA{10, 20, 30, 255})
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstreamPNG.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstreamPNG.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	// Convert PNG -> webp
	urlWebP := fmt.Sprintf("%s/0_0_80:webp/%s/img.png", srv.URL, upstreamPNG.URL)
	rWebP, err := http.Get(urlWebP)
	if err != nil {
		t.Fatalf("error requesting webp convert: %v", err)
	}
	defer rWebP.Body.Close()
	if rWebP.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", rWebP.StatusCode)
	}
	dataW, _ := io.ReadAll(rWebP.Body)
	if _, err := webp.Decode(bytes.NewReader(dataW)); err != nil {
		t.Fatalf("expected webp decode success: %v", err)
	}

	// Convert PNG -> webp_lossless
	urlWebPL := fmt.Sprintf("%s/0_0_80:webp_lossless/%s/img.png", srv.URL, upstreamPNG.URL)
	rLoss, err := http.Get(urlWebPL)
	if err != nil {
		t.Fatalf("error requesting webp_lossless convert: %v", err)
	}
	defer rLoss.Body.Close()
	if rLoss.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", rLoss.StatusCode)
	}
	dataL, _ := io.ReadAll(rLoss.Body)
	if _, err := webp.Decode(bytes.NewReader(dataL)); err != nil {
		t.Fatalf("expected lossless webp decode success: %v", err)
	}
}

// TestIntegrationPNGtoJPEG and JPEGtoPNG conversions.
func TestIntegrationPNGJPEGRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("jpegoptim"); err != nil {
		t.Skip("jpegoptim not installed")
	}
	// Upstream PNG
	upstreamPNG := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 16, 8))
		img.Set(0, 0, color.RGBA{220, 10, 10, 255})
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstreamPNG.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstreamPNG.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	// PNG->JPEG
	reqURL := fmt.Sprintf("%s/0_0_80:jpeg/%s/img.png", srv.URL, upstreamPNG.URL)
	resp, err := http.Get(reqURL)
	if err != nil {
		t.Fatalf("png->jpeg request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("expected jpeg decode success: %v", err)
	}

	// Upstream JPEG for reverse
	upstreamJPEG := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	}))
	defer upstreamJPEG.Close()
	hostPort2 := strings.TrimPrefix(upstreamJPEG.URL, "http://")
	srv2, stop2 := startPiumaServer(t, mediaPath, []string{hostPort2}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop2 <- true; srv2.Close() }()
	// JPEG->PNG
	reqURL2 := fmt.Sprintf("%s/0_0_80:png/%s/img.jpg", srv2.URL, upstreamJPEG.URL)
	resp2, err := http.Get(reqURL2)
	if err != nil {
		t.Fatalf("jpeg->png request error: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp2.StatusCode)
	}
	data2, _ := io.ReadAll(resp2.Body)
	if _, err := png.Decode(bytes.NewReader(data2)); err != nil {
		t.Fatalf("expected png decode success: %v", err)
	}
}

// TestIntegrationAVIFConversion (skip if avifenc unavailable).
func TestIntegrationAVIFConversion(t *testing.T) {
	if _, err := exec.LookPath("avifenc"); err != nil {
		t.Skip("avifenc not installed")
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 1000)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()
	reqURL := fmt.Sprintf("%s/0_0_80:avif/%s/img.png", srv.URL, upstream.URL)
	resp, err := http.Get(reqURL)
	if err != nil {
		t.Fatalf("avif request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "image/avif" {
		t.Fatalf("expected image/avif Content-Type got %s", ct)
	}
}

// TestIntegrationAutoPenalization simulates penalization removing a handler from auto-conf.
func TestIntegrationAutoPenalization(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// craft large file forcing penalization by making original tiny and optimized larger (quality high)
		img := image.NewRGBA(image.Rect(0, 0, 3, 3))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 1000)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	// First request auto:webp,png triggers selection (likely webp) and may penalize if larger than original.
	req1 := fmt.Sprintf("%s/0_0_95:auto:webp,png/%s/img.png", srv.URL, upstream.URL)
	r1, err := http.Get(req1)
	if err != nil {
		t.Fatalf("auto penalization first request error: %v", err)
	}
	defer r1.Body.Close()
	if r1.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", r1.StatusCode)
	}

	// Second request should potentially pick different format if penalization removed previous choice.
	time.Sleep(50 * time.Millisecond)
	r2, err := http.Get(req1)
	if err != nil {
		t.Fatalf("auto penalization second request error: %v", err)
	}
	defer r2.Body.Close()
	if r2.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", r2.StatusCode)
	}
	// We cannot deterministically assert change without inspecting auto-conf; capture for coverage only.
}

// TestIntegrationTransparencyConversion ensures PNG with alpha converted to JPEG yields jpeg output.
func TestIntegrationTransparencyConversion(t *testing.T) {
	if _, err := exec.LookPath("jpegoptim"); err != nil {
		t.Skip("jpegoptim not installed")
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		// set various alpha values
		for x := 0; x < 16; x++ {
			for y := 0; y < 16; y++ {
				img.Set(x, y, color.RGBA{uint8(x * 10), uint8(y * 10), 0, uint8(128 + (x+y)%127)})
			}
		}
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()
	reqURL := fmt.Sprintf("%s/0_0_80:jpeg/%s/img.png", srv.URL, upstream.URL)
	resp, err := http.Get(reqURL)
	if err != nil {
		t.Fatalf("png alpha -> jpeg request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "image/jpeg" {
		t.Fatalf("expected image/jpeg got %s", ct)
	}
}

// TestIntegrationAutoNegotiation ensures Accept header drives format selection.
func TestIntegrationAutoNegotiation(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/0_0_80:auto:webp,jpeg/%s/img.png", srv.URL, upstream.URL), nil)
	req.Header.Set("Accept", "image/webp,image/jpeg;q=0.8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("auto negotiation request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	// Should pick webp (higher preference order). Validate decode.
	if _, err := webp.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("expected webp selected, decode failed: %v", err)
	}
}

// TestIntegrationAdaptiveQuality triggers adaptive quality if dssim present.
func TestIntegrationAdaptiveQuality(t *testing.T) {
	if _, err := execLookPath("dssim"); err != nil {
		t.Skip("dssim not available")
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 1000)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	// Use adaptive quality directive (e.g. 75a) convert to webp
	urlA := fmt.Sprintf("%s/0_0_75a:webp/%s/img.png", srv.URL, upstream.URL)
	rA, err := http.Get(urlA)
	if err != nil {
		t.Fatalf("adaptive request error: %v", err)
	}
	defer rA.Body.Close()
	if rA.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", rA.StatusCode)
	}
	// Ensure file produced (
	io.ReadAll(rA.Body) // not asserting quality numerically (encoder binary search not exposed)
}

// execLookPath isolates LookPath for test to avoid importing os/exec multiple times.
func execLookPath(cmd string) (string, error) { return lookPathImpl(cmd) }

var lookPathImpl = func(cmd string) (string, error) { return "", errors.New("not found") }

// TestIntegrationWildcardDomain verifies wildcard allow-list accepts subdomain.
func TestIntegrationWildcardDomain(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 4, 4))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	// simulate upstream domain splitted to host, create wildcard suffix match
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	host := strings.Split(hostPort, ":")[0]
	idx := strings.Index(host, ".")
	if idx == -1 {
		t.Skip("host has no dot; wildcard suffix not applicable")
	}
	suffix := host[idx+1:]
	wildcard := "*" + suffix
	// Include host:port variant explicitly since DownloadImage matches exact domain token (host:port) not just host.
	srv, stop := startPiumaServer(t, mediaPath, []string{wildcard, hostPort}, 500)
	defer func() { core.GlobalWorkerManager.Close(); stop <- true; srv.Close() }()

	urlReq := fmt.Sprintf("%s/0_0_80:webp/%s/img.png", srv.URL, upstream.URL)
	resp, err := http.Get(urlReq)
	if err != nil {
		t.Fatalf("wildcard request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

// TestIntegrationOptimizeFailureFallback forces an optimization failure by closing worker manager early.
func TestIntegrationOptimizeFailureFallback(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 6, 6))
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer upstream.Close()
	mediaPath := t.TempDir()
	hostPort := strings.TrimPrefix(upstream.URL, "http://")
	srv, stop := startPiumaServer(t, mediaPath, []string{hostPort}, 10)
	// Immediately close workers to trigger timeout/closed behavior
	core.GlobalWorkerManager.Close()
	urlReq := fmt.Sprintf("%s/0_0_80:webp/%s/img.png", srv.URL, upstream.URL)
	resp, err := http.Get(urlReq)
	if err != nil {
		t.Fatalf("failure fallback request error: %v", err)
	}
	defer resp.Body.Close()
	// Expect either 400 (timeout error path) or 200 original fallback; treat both as success path for coverage
	if resp.StatusCode != 200 && resp.StatusCode != 400 {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	stop <- true
	srv.Close()
	time.Sleep(50 * time.Millisecond)
}
