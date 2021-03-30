package core

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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

func Dispatch(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	responseType := response.Header.Get("Content-Type")
	size := response.Header.Get("Content-Length")
	lastModified := response.Header.Get("Last-Modified")

	// Get Hash Name
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprint(imageParameters.prepareHashData(), response.Request.URL.String(), responseType, size, lastModified)))
	newFileName := base64.URLEncoding.EncodeToString(hash.Sum(nil))

	newImageTempPath := filepath.Join(options.PathTemp, newFileName)
	newImageRealPath := filepath.Join(options.PathMedia, newFileName)

	// Check if file exists
	if _, err := os.Stat(newImageRealPath); err == nil {
		var imageHandler ImageHandler
		if imageParameters.Convert != "" {
			imageHandler, err = NewImageHandlerByExtension(imageParameters.Convert)
			if err != nil {
				return "", "", err
			}
		} else {
			imageHandler, err = NewImageHandler(responseType)
			if err != nil {
				return "", "", err
			}
		}

		return newImageRealPath, imageHandler.ImageType(), nil
	}

	if _, loaded := FileMutex.LoadOrStore(newImageTempPath, true); loaded {
		return "", "", errors.New("Still elaborating")
	} else {
		img, err := os.Create(newImageTempPath)
		defer img.Close()
		if err != nil {
			return "", "", err
		}
		var buf bytes.Buffer
		copy := io.TeeReader(response.Body, &buf)
		io.Copy(img, copy)
		response.Body = io.NopCloser(&buf)
	}

	newOptions := options
	newOptions.PathTemp = newImageTempPath
	newOptions.PathMedia = newImageRealPath

	return asyncOptimize(response, imageParameters, newOptions)
}

func DownloadImage(originalUrl string) (*http.Response, error) {
	response, err := http.Get(originalUrl)
	if err != nil {
		return nil, errors.New("Error downloading file " + originalUrl)
	}
	return response, nil
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
