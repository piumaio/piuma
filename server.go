package main

import (
    "flag"
    "fmt"
    "github.com/julienschmidt/httprouter"
    "github.com/lotrekagency/piuma/core"
    "io"
    "log"
    "net/http"
    "os"
    "os/user"
    "path/filepath"
)

var pathtemp string
var pathmedia string

func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    var contentType string
    var response *http.Response
    imageURL := ps.ByName("url")[1:]
    parameters := ps.ByName("parameters")

    imageParameters, err := core.Parser(parameters)
    if err != nil {
        log.Printf("[ERROR]: parsing parameters [ %s ] : [ %s ]\n", parameters, err)
    }

    img, contentType, err := core.Optimize(imageURL, imageParameters, pathtemp, pathmedia)
    if err != nil {
        fmt.Printf("[ERROR]: optimizing image [ %s ]\n", err)
    } else {
        err = core.BuildResponse(w, img, contentType)
    }

    if err != nil {
        response, err = http.Get(imageURL)
        if err != nil {
            log.Printf("[ERROR]: downloading file [ %s ] - [ %s ]\n", imageURL, err)
        } else {
            var reader io.Reader = response.Body
            contentType = response.Header.Get("Content-Type")
            w.Header().Set("Content-Type", contentType) // <-- set the content-type header
            io.Copy(w, reader)
        }
    }
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

    flag.Parse()

    pathtemp = filepath.Join(pathmedia, "temp")

    os.MkdirAll(pathtemp, os.ModePerm)
    os.MkdirAll(pathmedia, os.ModePerm)

    router := httprouter.New()
    router.GET("/:parameters/*url", Manager)
    log.Fatal(http.ListenAndServe(":"+port, router))
}
