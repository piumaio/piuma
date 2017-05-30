package main

import (
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "./core"
    "fmt"
    //website "./website"
)

// func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//     fmt.Fprint(w, "hi man")
// }


func Manager(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // mi e arrivata una url, la passo a parser, cosicchè mi possa tornare un
    // prendo il parser array e lo do ad optimizer che mi ridà la path di una img
    // prendo il path della img, e lo sparo sulla response così
    //core.Parser()
    img, content_type, err := core.Optimize(ps.ByName("url")[1:], 200, 0, 80)
    if err != nil {
       fmt.Println(err)
    }
    core.BuildResponse(w, img, content_type)
}

func main() {
    router := httprouter.New()
    router.GET("/:parameters/*url", Manager)
    log.Fatal(http.ListenAndServe(":8080", router))
    // img, err := core.Optimize("http://tvl.lotrek.it/media/MRIM_02_ok.jpg", 500, 0, 80)
    // if err != nil {
    //    log.Fatal(err)
    // }
    // fmt.Println(img)
    // img, err = core.Optimize("https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png", 200, 0, 80)
    // if err != nil {
    //    log.Fatal(err)
    // }
    // fmt.Println(img)
}
