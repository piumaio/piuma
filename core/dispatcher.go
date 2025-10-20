package core

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// OptimizationResult holds the path to the optimized image, its mime type and
// any error produced during processing. (Currently unused; retained for future
// synchronous API refactors.)
type OptimizationResult struct {
	image_path, mime_type string
	err                   error
}

// Options bundles runtime paths and timeout controls for processing:
//
//	PathTemp  -> temporary file containing unoptimized original bytes
//	PathMedia -> final destination path for optimized artifact
//	Timeout   -> max time (milliseconds) to wait for async optimization
//	             before returning a timeout error to caller
type Options struct {
	PathTemp, PathMedia string
	Timeout             int
}

// FileMutex prevents concurrent optimization of the same temp file (hash) to
// avoid duplicated CPU work and race conditions writing the output.
var FileMutex sync.Map

// HttpCacheMutex stores expiration timestamps for cached HTTP responses used
// by DownloadImage; the purge goroutine periodically sweeps expired entries.
var HttpCacheMutex sync.Map

// Dispatch orchestrates format auto-selection (when requested) and prepares
// file system state before delegating actual optimization to async workers.
// It returns the final optimized file path, mime type, or an error if still
// elaborating or setup fails. For existing cached optimized files it short
// circuits and returns instantly.
func Dispatch(request *http.Request, response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	if strings.HasPrefix(imageParameters.Convert, "auto") {
		autoConfPath := filepath.Join(options.PathMedia, imageParameters.GenerateHash(response))
		preferredConverts := []string{}
		if strings.HasPrefix(imageParameters.Convert, "auto:") {
			preferredConverts = strings.Split(strings.Split(imageParameters.Convert, ":")[1], ",")
		}
		imageHandler, err := AutoImageHandler(request, response, autoConfPath, preferredConverts)
		if err != nil {
			return "", "", err
		}
		imageParameters.Convert = imageHandler.ImageExtension()
	}

	newFileName := imageParameters.GenerateHash(response)

	newImageTempPath := filepath.Join(options.PathTemp, newFileName)
	newImageRealPath := filepath.Join(options.PathMedia, newFileName)

	// Check if file exists
	if file, err := os.Open(newImageRealPath); err == nil {
		defer file.Close()
		imageHandler, err := NewImageHandlerByBytes(file)
		if err == nil {
			return newImageRealPath, imageHandler.ImageType(), nil
		}
	}

	if _, loaded := FileMutex.LoadOrStore(newImageTempPath, true); loaded {
		return "", "", errors.New("Still elaborating")
	} else {
		img, err := os.Create(newImageTempPath)
		if err != nil {
			return "", "", err
		}
		var buf bytes.Buffer
		copy := io.TeeReader(response.Body, &buf)
		_, err = io.Copy(img, copy)
		if err != nil {
			return "", "", err
		}
		response.Body = io.NopCloser(&buf)
		img.Close()
	}

	newOptions := options
	newOptions.PathTemp = newImageTempPath
	newOptions.PathMedia = newImageRealPath

	return asyncOptimize(response, imageParameters, newOptions)
}

// DownloadImage retrieves the remote image enforcing an allow-list of domains
// (supports wildcard prefix "*" for subdomains). It caches successful HTTP
// responses on disk for cacheDelay seconds. Returns a possibly non-200
// response alongside sentinel errors: invalid_domain, invalid_status_code,
// invalid_content_type. Callers may stream body on error for graceful
// degradation.
func DownloadImage(originalUrl string, cacheDelay int, allowed_domains []string) (*http.Response, error) {
	parts := strings.Split(originalUrl, "/")
	if len(parts) < 3 {
		return &http.Response{StatusCode: 400, Request: &http.Request{URL: &url.URL{}}}, errors.New("invalid_domain")
	}

	image_domain := parts[2]
	domain_is_valid := false

	for _, domain := range allowed_domains {
		if strings.HasPrefix(domain, "*") {
			if strings.HasSuffix(image_domain, strings.TrimLeft(domain, "*")) {
				domain_is_valid = true
				break
			}
		} else if domain == image_domain {
			domain_is_valid = true
			break
		}
	}

	if !domain_is_valid {
		request, _ := http.NewRequest("GET", originalUrl, nil)
		response := &http.Response{
			Request:    request,
			StatusCode: 403,
		}
		return response, errors.New("invalid_domain")
	}

	hash := sha1.New()
	hash.Write([]byte(originalUrl))
	filename := filepath.Join(os.TempDir(), "piuma_http_cache", base64.URLEncoding.EncodeToString(hash.Sum(nil)))

	if value, ok := HttpCacheMutex.Load(filename); ok && value.(int64) > time.Now().Unix() {
		cacheData, err := os.Open(filename)
		if err == nil {
			buffer := bufio.NewReader(cacheData)
			request, err := http.NewRequest("GET", originalUrl, nil)
			if err == nil {
				response, err := http.ReadResponse(buffer, request)

				if err == nil {
					return response, nil
				}
			}
		}
	}

	response, err := http.Get(originalUrl)
	if err != nil {
		return nil, errors.New("Error downloading file " + originalUrl)
	}
	if response.StatusCode != 200 {
		return response, errors.New("invalid_status_code")
	}

	if strings.Split(response.Header.Get("Content-Type"), "/")[0] != "image" {
		return response, errors.New("invalid_content_type")
	}

	cacheData, err := httputil.DumpResponse(response, true)
	if err != nil {
		return response, nil
	}
	err = os.WriteFile(filename, cacheData, 0644)
	if err != nil {
		return response, nil
	}

	HttpCacheMutex.Store(filename, time.Now().Unix()+int64(cacheDelay))

	return response, nil
}

// StartHttpCachePurge launches a goroutine that periodically scans cached
// response entries and deletes expired files. Returns a channel used to signal
// termination of the purge loop.
func StartHttpCachePurge(checkIntervalSeconds int) chan bool {
	ticker := time.NewTicker(time.Duration(checkIntervalSeconds) * time.Second)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				HttpCacheMutex.Range(func(key, value interface{}) bool {
					if value.(int64) < time.Now().Unix() {
						os.Remove(key.(string))
					}
					return true
				})
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

// BuildResponse writes the optimized file contents to the response writer and
// sets the Content-Type header. On file read failure returns an error and the
// caller may fallback to original bytes or error JSON.
func BuildResponse(w http.ResponseWriter, imagePath string, contentType string) error {
	img, err := os.Open(imagePath)
	if err != nil {
		return errors.New("error reading from optimized file")
	}
	defer img.Close()
	w.Header().Set("Content-Type", contentType) // <-- set the content-type header
	io.Copy(w, img)
	return nil
}

// IsImage checks if the provided HTTP response represents an image based on
// its Content-Type header prefix.
func IsImage(response *http.Response) bool {
	return strings.Contains(response.Header.Get("Content-Type"), "image/")
}
