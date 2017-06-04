package main

import (
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "./core"
    "fmt"
    "io"
)

func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    width, height, quality, err := core.Parser(ps.ByName("parameters"))
    if err == nil {
        fmt.Println(err)
        img, content_type, err := core.Optimize(ps.ByName("url")[1:], width, height, quality)
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
    router := httprouter.New()
    router.GET("/:parameters/*url", Manager)
    log.Fatal(http.ListenAndServe(":8080", router))
}
