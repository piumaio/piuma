package core

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

// GlobalWorkerManager is a process-wide singleton used by Dispatch/asyncOptimize
// to queue optimization jobs. Created in main() and shut down on server exit.
var GlobalWorkerManager *WorkerManager

// OptimizeFunc is an overridable seam used by tests to inject latency or
// faults into the optimization pipeline for deterministic concurrency
// behaviors. Defaults to Optimize.
var OptimizeFunc = Optimize

// output represents the result of an optimization job returned over a channel.
// When err != nil path/mime may be empty. Channel will be closed after send.
type output struct {
	path string
	mime string
	err  error
}

// input encloses all required data for a worker to perform optimization. It
// includes the original HTTP response (with body readable), transformation
// directives, runtime paths and a one-element buffered channel to publish its
// output asynchronously.
type input struct {
	response        *http.Response
	imageParameters *ImageParameters
	options         *Options
	result          chan output
}

// WorkerManager coordinates a pool of goroutines consuming optimization
// requests. Backpressure is applied via a bounded buffered channel. When the
// channel is full Dispatch returns nil causing the caller to treat it as a
// timeout scenario. Close() gracefully drains workers.
type WorkerManager struct {
	closed    bool
	data      chan input
	waitGroup *sync.WaitGroup
	close     chan bool
}

// NewWorkerManager builds a Manager with a pre-sized input buffer so bursts of
// requests can enqueue without immediate blocking (up to capacity).
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		data:      make(chan input, 1024),
		waitGroup: &sync.WaitGroup{},
		close:     make(chan bool),
		closed:    false,
	}
}

// asyncOptimize submits an optimization job and waits for its result or a
// timeout (if options.Timeout > 0). Errors are surfaced transparently; on
// timeout it returns a "Timed out" sentinel error enabling a fallback path.
func asyncOptimize(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	workerResponse := GlobalWorkerManager.Dispatch(response, imageParameters, options)
	if options.Timeout != 0 && workerResponse != nil {
		select {
		case result := <-workerResponse:
			return result.path, result.mime, result.err
		case <-time.After(time.Duration(options.Timeout) * time.Millisecond):
			return "", "", errors.New("timed out")
		}
	}
	return "", "", errors.New("timed out")
}

// Dispatch attempts to enqueue a new job returning a channel for its eventual
// output. If the internal buffer is saturated or the manager is closed it
// returns nil, signaling to callers that they should enforce timeout behavior.
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

// Run starts a single worker goroutine that loops reading from the input
// channel until a close signal is received. Each request triggers Optimize,
// after which locks and logging are handled before publishing the result.
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
				path, mime, err := OptimizeFunc(req.response, req.imageParameters, req.options)
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

// Close signals workers to terminate and waits until all have exited.
// Subsequent Dispatch calls will return nil.
func (w *WorkerManager) Close() {
	if !w.closed {
		close(w.close)
		w.closed = true
		w.waitGroup.Wait()
	}
}
