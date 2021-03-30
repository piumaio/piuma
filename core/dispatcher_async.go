package core

import (
	"errors"
	"log"
	"net/http"
	"time"
)

func asyncOptimize(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	if options.Timeout == 0 {
		return asyncOptimizeNoTimeout(response, imageParameters, options)
	}
	c := make(chan OptimizationResult)
	go func(r *http.Response, imP *ImageParameters, o *Options) {
		defer close(c)
		path, mime, err := Optimize(response, imP, o)
		FileMutex.Delete(o.PathTemp)
		c <- OptimizationResult{path, mime, err}
		if err == nil {
			log.Printf("[INFO] Done with %s \n", r.Request.URL)
		} else {
			log.Printf("[ERROR] [%s] %s \n", r.Request.URL, err)
		}
	}(response, imageParameters, options)
	select {
	case result := <-c:
		return result.image_path, result.mime_type, result.err
	case <-time.After(time.Duration(options.Timeout) * time.Millisecond):
		return "", "", errors.New("Timed out")
	}
}

func asyncOptimizeNoTimeout(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	go func(r *http.Response, imP *ImageParameters, o *Options) {
		_, _, err := Optimize(response, imP, o)
		FileMutex.Delete(o.PathTemp)
		if err == nil {
			log.Printf("[INFO] Done with %s \n", r.Request.URL)
		} else {
			log.Printf("[ERROR] [%s] %s \n", r.Request.URL, err)
		}
	}(response, imageParameters, options)
	return "", "", errors.New("Timed out")
}
