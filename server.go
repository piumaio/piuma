package main

import (
    "github.com/julienschmidt/httprouter"
    "github.com/lotrekagency/piuma/core"
    "path/filepath"
    "net/http"
    "log"
    "fmt"
    "io"
    "os"
    "os/user"
)

var pathtemp string = ""
var pathmedia string = ""

func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    width, height, quality, err := core.Parser(ps.ByName("parameters"))
    if err == nil {
        img, content_type, err := core.Optimize(ps.ByName("url")[1:], width, height, quality, pathtemp, pathmedia)
        if err != nil {
            fmt.Println(err)
        }
        core.BuildResponse(w, img, content_type)
    }

    if err != nil {
        response, err := http.Get(ps.ByName("url")[1:])
        if err != nil {
           fmt.Println("Error downloading file " + ps.ByName("url")[1:])
        } else {
            var reader io.Reader = response.Body
            content_type := response.Header.Get("Content-Type")
            w.Header().Set("Content-Type", content_type) // <-- set the content-type header
            io.Copy(w, reader)
        }
    }
}

func main() {

    usr, err := user.Current()

    if err != nil {
        fmt.Println(err)
        return
    }

    pathtemp = filepath.Join(usr.HomeDir, ".piuma", "temp")
    pathmedia = filepath.Join(usr.HomeDir, ".piuma", "media")

    os.MkdirAll(pathtemp, os.ModePerm)
    os.MkdirAll(pathmedia, os.ModePerm)

    router := httprouter.New()
    router.GET("/:parameters/*url", Manager)
    log.Fatal(http.ListenAndServe(":8080", router))
}
