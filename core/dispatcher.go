package core

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type OptimizationResult struct {
	image_path, mime_type string
	err                   error
}

type Options struct {
	PathTemp, PathMedia string
	Timeout             int
}

var FileMutex sync.Map
var HttpCacheMutex sync.Map

func Dispatch(request *http.Request, response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	responseType := response.Header.Get("Content-Type")
	size := response.Header.Get("Content-Length")
	lastModified := response.Header.Get("Last-Modified")

	if imageParameters.Convert == "auto" {
		imageHandler, err := AutoImageHandler(request, response)
		if err != nil {
			return "", "", err
		}
		imageParameters.Convert = imageHandler.ImageExtension()
	}

	// Get Hash Name
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprint(imageParameters.PrepareHashData(), response.Request.URL.String(), responseType, size, lastModified)))
	newFileName := base64.URLEncoding.EncodeToString(hash.Sum(nil))

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

func DownloadImage(originalUrl string, cacheDelay int) (*http.Response, error) {
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
	cacheData, err := httputil.DumpResponse(response, true)
	if err != nil {
		return response, nil
	}
	err = ioutil.WriteFile(filename, cacheData, 0644)
	if err != nil {
		return response, nil
	}

	HttpCacheMutex.Store(filename, time.Now().Unix()+int64(cacheDelay))

	return response, nil
}

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

func BuildResponse(w http.ResponseWriter, imagePath string, contentType string) error {
	img, err := os.Open(imagePath)
	if err != nil {
		return errors.New("Error reading from optimized file")
	}
	defer img.Close()
	w.Header().Set("Content-Type", contentType) // <-- set the content-type header
	io.Copy(w, img)
	return nil
}
