package core

import (
    "net/http"
    "io"
    "os"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "image/jpeg"
    //"image/png"
    "github.com/nfnt/resize"
    "errors"
)

func Optimize(original_url string, width uint, height uint, quality uint) (string, error) {

    // Get Hash Name
    hash := sha1.New()
    hash.Write([]byte(fmt.Sprint(width, height, quality, original_url)))
    new_file_name := base64.URLEncoding.EncodeToString(hash.Sum(nil))
    fmt.Println(new_file_name)

    new_image_temp_path := "temp/" + new_file_name

    // Download file
    response, err := http.Get(original_url)
    if err != nil {
       return "", errors.New("Error downloading file " + original_url)
    }

    // Decode and resize
    var r io.Reader = response.Body

    img, err := jpeg.Decode(r)
    if err != nil {
        return "", errors.New("Error decoding response")
    }

    new_image := resize.Resize(width, height, img, resize.NearestNeighbor)

    new_file_img, err := os.Create(new_image_temp_path)
    if err != nil {
        return "", errors.New("Error creating new image")
    }
    defer new_file_img.Close()

    // Encode new image
    err = jpeg.Encode(new_file_img, new_image, nil)
    if err != nil {
        return "", errors.New("Error encoding response")
    }

    return "asd", nil
}
