package main

import (
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "fmt"
    core "./core"
    //website "./website"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "hi man")
}


func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // mi e arrivata una url, la passo a parser, cosicchè mi possa tornare un
    // prendo il parser array e lo do ad optimizer che mi ridà la path di una img
    // prendo il path della img, e lo sparo sulla response così
    core.Parser()
}

func main() {
   router := httprouter.New()
   router.GET("/:url", Manager)
   router.GET("/statics", Index)
   log.Fatal(http.ListenAndServe(":8080", router))
}
