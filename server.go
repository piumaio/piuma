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

    imageParameters, err := core.Parser(ps.ByName("parameters"))
    if err == nil {
        img, contentType, err := core.Optimize(ps.ByName("url")[1:], imageParameters, pathtemp, pathmedia)
        if err != nil {
            fmt.Println(err)
        } else {
            err = core.BuildResponse(w, img, contentType)
        }
    }

    if err != nil {
        response, err = http.Get(ps.ByName("url")[1:])
        if err != nil {
            fmt.Println("Error downloading file " + ps.ByName("url")[1:])
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
        log.Printf("[ERROR]: failed getting user [ %s ]", err)
        os.Exit(1)
    }

    // TODO: This could be passed in as an argument and/or read
    // from an enviroment variable
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
