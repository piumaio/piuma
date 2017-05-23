package piuma

import (
    "net/http"
    "io"
    "os"
    "crypto/sha1"
    "strings"
    "image"
    "image/jpeg"
    "image/png"
    "github.com/nfnt/resize"
    "log"
    "strconv"
)

//dimensions, quality, url
func Recognize(data [3]string) string{
    //Controllo se ho fatto questa modifica
    done_check:=[2]string {}
    done_check[0],done_check[1]=RequestDone(data)
    hf := done_check[0]
    if done_check[1] == "found" {
        return "/img/"+hf
    }

    //Crea una copia locale del file originale
    file_split := strings.Split(data[2],"/")
    src, err := os.Create("/img/original/"+file_split[len(file_split)-1])
    if err != nil {
        log.Fatal(err)
        return "err"
    }
    defer src.Close()

    //Apre il file originale
    image_original, err := http.Get(data[2])
    if err != nil {
        log.Fatal(err)
        return "err"
    }
    defer image_original.Body.Close()

    //Copia il file originale nella copia locale
    _, err = io.Copy(src, image_original.Body)
    if err != nil {
        log.Fatal(err)
        return "err"
    }

    new_dimensions := strings.Split(data[0],"x")
    
    new_width, err:=strconv.Atoi(new_dimensions[0])
    if err != nil {
        log.Fatal(err)
        return "err"
    }
    new_height, err:=strconv.Atoi(new_dimensions[1])
    if err != nil {
        log.Fatal(err)
        return "err"
    }

    out, err := os.Open("/img/"+hf)
    if err != nil {
        log.Fatal(err)
        return "err"
    }
    defer out.Close()

    //Decodifico il file come immagine
    if done_check[1] == "jpeg" || done_check[1] == "jpg" {
        dec_src, _, err := image.Decode(src)
        if err != nil {
            log.Fatal(err)
            return "err"
        }

        new_image := resize.Resize(uint(new_width), uint(new_height), dec_src, resize.NearestNeighbor)

        //Codfica la nuova immagine nel file
        err = jpeg.Encode(out, new_image, nil)
        if err != nil {
            log.Fatal(err)
            return "err"
        }

    }else{
        dec_src, _, err := image.Decode(src)
        if err != nil {
            log.Fatal(err)
            return "err"
        }

        new_image := resize.Resize(uint(new_width), uint(new_height), dec_src, resize.NearestNeighbor)

        //Codfica la nuova immagine nel file
        err = png.Encode(out, new_image)
        if err != nil {
            log.Fatal(err)
            return "err"
        }
    }

    return "/img/"+hf
}

func RequestDone(data [3]string) (string,string){
    //Preparo il corretto nome del file
    url_split := strings.Split(data[2],"/")
    file_split := strings.Split(url_split[len(url_split)-1],".")
    hashable := data[0]+data[1]+data[2]
    h := sha1.New()
    h.Write([]byte(hashable))
    hf := string(h.Sum(nil))
    hf = hf+"."+file_split[len(file_split)-1]

    //Se il file esiste lo restituisco, altrimenti lo creo e lo restituisco
    if _, err := os.Stat("/img/"+hf); !os.IsNotExist(err) {
        return hf, "found"
    }else{
        out, err := os.Create(hf)
        if err != nil  {
            log.Fatal(err)
            return "err","err"
        }
        defer out.Close()
        return hf, file_split[len(file_split)-1]
    }
}