package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/piumaio/piuma/core"
)

var pathtemp string
var pathmedia string
var timeout int

func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var contentType string
	imageURL := ps.ByName("url")[1:]
	parameters := ps.ByName("parameters")

	imageParameters, err := core.Parser(parameters)
	if err != nil {
		log.Printf("[ERROR]: parsing parameters [ %s ] : [ %s ]\n", parameters, err)
		return
	}

	image, err := core.DownloadImage(imageURL)
	if err != nil {
		log.Printf("[ERROR]: error while downloading image [ %s ]\n", err)
		return
	}

	img, contentType, err := core.Dispatch(image, &imageParameters, &core.Options{pathtemp, pathmedia, timeout})
	if err != nil {
		if err.Error() != "Timed out" {
			fmt.Printf("[ERROR]: optimizing image [ %s ]\n", err)
		}
	} else {
		err = core.BuildResponse(w, img, contentType)
	}

	if err != nil {
		contentType = image.Header.Get("Content-Type")
		w.Header().Set("Content-Type", contentType) // <-- set the content-type header
		io.Copy(w, image.Body)
	}

	image.Body.Close()
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Printf("[ERROR]: failed getting user [ %s ]\n", err)
		os.Exit(1)
	}

	var port = "8080"

	flag.StringVar(&port, "port", port, "Port where piuma will run")
	flag.StringVar(&pathmedia, "mediapath", filepath.Join(usr.HomeDir, ".piuma", "media"), "Media path")
	flag.IntVar(&timeout, "timeout", 0, "Maximum time to wait for image elaboration")

	flag.Parse()

	pathtemp = filepath.Join(pathmedia, "temp")

	os.MkdirAll(pathtemp, os.ModePerm)
	os.MkdirAll(pathmedia, os.ModePerm)

	router := httprouter.New()
	router.GET("/:parameters/*url", Manager)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
