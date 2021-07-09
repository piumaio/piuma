package core

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

var GlobalWorkerManager *WorkerManager

type output struct {
	path string
	mime string
	err  error
}

type input struct {
	response        *http.Response
	imageParameters *ImageParameters
	options         *Options
	result          chan output
}

type WorkerManager struct {
	closed    bool
	data      chan input
	waitGroup *sync.WaitGroup
	close     chan bool
}

func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		data:      make(chan input, 1024),
		waitGroup: &sync.WaitGroup{},
		close:     make(chan bool),
		closed:    false,
	}
}

func asyncOptimize(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	workerResponse := GlobalWorkerManager.Dispatch(response, imageParameters, options)
	if options.Timeout != 0 && workerResponse != nil {
		select {
		case result := <-workerResponse:
			return result.path, result.mime, result.err
		case <-time.After(time.Duration(options.Timeout) * time.Millisecond):
			return "", "", errors.New("Timed out")
		}
	}
	return "", "", errors.New("Timed out")
}

func (w *WorkerManager) Dispatch(response *http.Response, imageParameters *ImageParameters, options *Options) chan output {
	if !w.closed {
		output := make(chan output, 1)
		select {
		case w.data <- input{response: response, imageParameters: imageParameters, options: options, result: output}:
			return output
		default:
			return nil
		}
	}
	return nil
}

func (w *WorkerManager) Run() {
	w.waitGroup.Add(1)

	go func() {
		for {
			select {
			case <-w.close:
				w.closed = true
				w.waitGroup.Done()
				return
			case req := <-w.data:
				path, mime, err := Optimize(req.response, req.imageParameters, req.options)
				FileMutex.Delete(req.options.PathTemp)
				req.result <- output{path, mime, err}
				close(req.result)
				if err != nil {
					log.Printf("[ERROR] [%s] %s \n", req.response.Request.URL, err)
				} else {
					log.Printf("[INFO] Done with %s \n", req.response.Request.URL)
				}
			}
		}
	}()

}

func (w *WorkerManager) Close() {
	w.close <- true
	w.waitGroup.Wait()
	close(w.close)
}
