package core

import (
    "net/http"
    "io"
    "os"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "github.com/nfnt/resize"
    "errors"
)


func Optimize(original_url string, width uint, height uint, quality uint) (string, string, error) {

    // Download file
    response, err := http.Get(original_url)
    if err != nil {
       return "", "", errors.New("Error downloading file " + original_url)
    }

    defer response.Body.Close()

    // Detect image type, size and last modified
    response_type := response.Header.Get("Content-Type")
    size := response.Header.Get("Content-Length")
    last_modified := response.Header.Get("Last-Modified")
    fmt.Println(response_type)
    fmt.Println(response.Header)

    // Get Hash Name
    hash := sha1.New()
    hash.Write([]byte(fmt.Sprint(width, height, quality, original_url, response_type, size, last_modified)))
    new_file_name := base64.URLEncoding.EncodeToString(hash.Sum(nil))
    fmt.Println(new_file_name)

    new_image_temp_path := "temp/" + new_file_name

    // Check if file exists
    if _, err := os.Stat(new_image_temp_path); err == nil {
      return new_image_temp_path, response_type, nil
    }

    // Decode and resize
    var reader io.Reader = response.Body
    var img image.Image

    if response_type == "image/jpeg" {
        img, err = jpeg.Decode(reader)
    } else if response_type == "image/png" {
        img, err = png.Decode(reader)
    } else {
        return "", "", errors.New("Format not supported")
    }

    if err != nil {
        return "", "", errors.New("Error decoding response")
    }

    new_image := resize.Resize(width, height, img, resize.NearestNeighbor)

    new_file_img, err := os.Create(new_image_temp_path)
    if err != nil {
        return "", "", errors.New("Error creating new image")
    }
    defer new_file_img.Close()

    // Encode new image
    if response_type == "image/jpeg" {
        err = jpeg.Encode(new_file_img, new_image, nil)
    } else if response_type == "image/png" {
        err = png.Encode(new_file_img, new_image)
    }
    if err != nil {
        return "", "", errors.New("Error encoding response")
    }

    return new_image_temp_path, response_type, nil
}


func BuildResponse (w http.ResponseWriter, image_path string, content_type string) (error){
    img, err := os.Open(image_path)
    if err != nil {
        return errors.New("Error reading from optimized file")
    }
    defer img.Close()
    w.Header().Set("Content-Type", content_type) // <-- set the content-type header
    io.Copy(w, img)
    return nil
}
