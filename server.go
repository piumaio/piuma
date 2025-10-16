package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/piumaio/piuma/core"
)

var pathtemp string
var pathmedia string
var timeout int
var httpCacheTTL int
var httpCachePurgeInterval int
var workers int
var version string
var domains string
var domains_list []string

// processImage is the main request handler for transformation requests.
// Route pattern: /:parameters/*url where :parameters encodes ImageParameters
// (see core.Parser) and *url is the upstream image URL to fetch. It performs:
//   - parameter parsing
//   - remote download (with domain allow-list and cache)
//   - async optimization dispatch
//   - response streaming or graceful fallback
//
// Errors during optimization fall back to original image when available.
func processImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var contentType string
	rawURL := ps.ByName("url")
	if len(rawURL) > 0 && rawURL[0] == '/' {
		rawURL = rawURL[1:]
	}
	imageURL := rawURL
	parameters := ps.ByName("parameters")

	imageParameters, err := core.Parser(parameters)
	if err != nil {
		log.Printf("[ERROR]: parsing parameters [ %s ] : [ %s ]\n", parameters, err)
		return
	}
	if len(domains_list) == 0 {
		domains_list = []string{r.Host}
	}
	image, err := core.DownloadImage(imageURL, httpCacheTTL, domains_list)
	if err != nil {
		writeError(w, image, err)
		log.Printf("[ERROR]: error while downloading image [ %s ]\n", err)
		return
	}

	img, contentType, err := core.Dispatch(r, image, &imageParameters, &core.Options{PathTemp: pathtemp, PathMedia: pathmedia, Timeout: timeout})
	if err != nil {
		if err.Error() != "Timed out" {
			fmt.Printf("[ERROR]: optimizing image [ %s ]\n", err)
		}
	} else {
		err = core.BuildResponse(w, img, contentType)
	}

	if err != nil {
		if image != nil {
			contentType = image.Header.Get("Content-Type")
			w.Header().Set("Content-Type", contentType) // <-- set the content-type header
			if image.Body != nil {
				io.Copy(w, image.Body)
			}
		}
	}

	if image != nil && image.Body != nil {
		image.Body.Close()
	}
}

// writeError converts internal sentinel errors into structured JSON responses
// with appropriate HTTP status codes. If the original upstream response is
// available its status/content-type may be surfaced for debugging.
func writeError(w http.ResponseWriter, r *http.Response, err error) {
	var data = map[string]interface{}{
		"error":  strings.ToUpper(err.Error()),
		"detail": "",
	}

	switch err.Error() {
	case "invalid_status_code":
		w.WriteHeader(http.StatusNotFound)
		if r != nil {
			data["detail"] = fmt.Sprintf("Original status code was: %d.", r.StatusCode)
		}
	case "invalid_content_type":
		w.WriteHeader(http.StatusUnsupportedMediaType)
		if r != nil {
			data["detail"] = fmt.Sprintf("Original Content-Type was: %s.", r.Header.Get("Content-Type"))
		}
	case "invalid_domain":
		w.WriteHeader(http.StatusForbidden)
		if r != nil && r.Request != nil && r.Request.URL != nil {
			data["detail"] = fmt.Sprintf("Images from domain %s are not allowed.", r.Request.URL.Host)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getInfo returns basic service metadata: supported extensions and version.
// It also handles CORS pre-flight OPTIONS requests.
func getInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == "OPTIONS" {
		return
	}

	var data = map[string]interface{}{
		"extensions": map[string]string{},
		"version":    version,
	}

	for _, v := range core.GetAllImageHandlers() {
		data["extensions"].(map[string]string)[v.ImageType()] = v.ImageExtension()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
}

// main bootstraps configuration (flags, paths, domain allow-list), starts
// worker pool and HTTP cache purge loop, then serves HTTP traffic. On server
// shutdown it ensures worker manager and purge goroutine terminate cleanly.
// Config holds runtime settings parsed from flags.
type Config struct {
	Port                   string
	MediaPath              string
	Timeout                int
	HTTPCacheTTL           int
	HTTPCachePurgeInterval int
	Workers                int
	Domains                string
}

// parseConfig builds a Config from the provided argument slice (excluding program name).
// It does not mutate global variables; Initialize applies Config to globals.
func parseConfig(args []string) (Config, error) {
	usr, err := user.Current()
	if err != nil {
		return Config{}, fmt.Errorf("getting current user: %w", err)
	}
	fs := flag.NewFlagSet("piuma", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // silence in tests
	cfg := Config{Port: "8080", MediaPath: filepath.Join(usr.HomeDir, ".piuma", "media"), Timeout: 0, HTTPCacheTTL: 3600, HTTPCachePurgeInterval: 3600, Workers: 4}
	fs.StringVar(&cfg.Port, "port", cfg.Port, "Port where piuma will run")
	fs.StringVar(&cfg.MediaPath, "mediapath", cfg.MediaPath, "Media path")
	fs.IntVar(&cfg.Timeout, "timeout", cfg.Timeout, "Maximum time to wait for image elaboration (in seconds)")
	fs.IntVar(&cfg.HTTPCacheTTL, "httpCacheTTL", cfg.HTTPCacheTTL, "Time To Live (in seconds) for HTTP Response Cache")
	fs.IntVar(&cfg.HTTPCachePurgeInterval, "httpCachePurgeInterval", cfg.HTTPCachePurgeInterval, "Interval for deleting unused cache (in seconds)")
	fs.IntVar(&cfg.Workers, "workers", cfg.Workers, "Number of workers to instantiate")
	fs.StringVar(&cfg.Domains, "domains", "", "Allowed domains, separated by commas (e.g. domain1.com,domain2.com)")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Initialize mutates global runtime settings, prepares directories, starts background
// purge loop and workers, and returns the router plus a shutdown function.
func Initialize(cfg Config) (*httprouter.Router, func(), error) {
	pathmedia = cfg.MediaPath
	timeout = cfg.Timeout
	httpCacheTTL = cfg.HTTPCacheTTL
	httpCachePurgeInterval = cfg.HTTPCachePurgeInterval
	workers = cfg.Workers
	domains = cfg.Domains

	if domains == "" {
		log.Printf("[WARNING]: No allowed domains specified, using the current domain")
		domains_list = []string{}
	} else {
		domains_list = strings.Split(domains, ",")
	}
	pathtemp = filepath.Join(pathmedia, "temp")
	if err := os.MkdirAll(pathtemp, os.ModePerm); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(pathmedia, os.ModePerm); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(filepath.Join(os.TempDir(), "piuma_http_cache"), os.ModePerm); err != nil {
		return nil, nil, err
	}

	router := httprouter.New()
	router.GET("/", getInfo)
	router.GET("/:parameters/*url", processImage)

	stopPurgeChan := core.StartHttpCachePurge(httpCachePurgeInterval)
	core.GlobalWorkerManager = core.NewWorkerManager()
	for i := 0; i < workers || i < 1; i++ {
		core.GlobalWorkerManager.Run()
	}
	shutdown := func() {
		core.GlobalWorkerManager.Close()
		stopPurgeChan <- true
	}
	return router, shutdown, nil
}

// run wires together configuration parsing and initialization returning server error (if any).
func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	router, shutdown, err := Initialize(cfg)
	if err != nil {
		return err
	}
	defer shutdown()
	log.Printf("Allowed domains: %s", domains)
	return http.ListenAndServe(":"+cfg.Port, router)
}

// main remains thin delegating to run for testability.
func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
