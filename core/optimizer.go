package core

import (
    "crypto/sha1"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/nfnt/resize"
    "image"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "sync"
)

func Optimize(originalUrl string, imageParameters ImageParameters, pathtemp string, pathmedia string) (string, string, error) {

    // Download file
    response, err := http.Get(originalUrl)
    if err != nil {
        return "", "", errors.New("Error downloading file " + originalUrl)
    }

    defer response.Body.Close()

    // Detect image type, size and last modified
    responseType := response.Header.Get("Content-Type")
    size := response.Header.Get("Content-Length")
    lastModified := response.Header.Get("Last-Modified")

    // Get Hash Name
    hash := sha1.New()
    hash.Write([]byte(fmt.Sprint(imageParameters.Width, imageParameters.Height, imageParameters.Quality, originalUrl, responseType, size, lastModified)))
    newFileName := base64.URLEncoding.EncodeToString(hash.Sum(nil))

    newImageTempPath := filepath.Join(pathtemp, newFileName)
    newImageRealPath := filepath.Join(pathmedia, newFileName)

    // Check if file exists
    if _, err := os.Stat(newImageRealPath); err == nil {
        return newImageRealPath, responseType, nil
    }

    // Decode and resize
    var reader io.Reader = response.Body
    var newFileImg *os.File
    var mu = &sync.Mutex{}

    mu.Lock()
    if _, err := os.Stat(newImageTempPath); err == nil {
        return "", "", errors.New("Still elaborating")
    }

    newFileImg, err = os.Create(newImageTempPath)
    mu.Unlock()

    var img image.Image
    var imageHandler ImageHandler

    imageHandler, err = NewImageHandler(responseType)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", err
    }

    img, err = imageHandler.Decode(reader)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", errors.New("Error decoding response")
    }

    newImage := resize.Resize(imageParameters.Width, imageParameters.Height, img, resize.NearestNeighbor)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", errors.New("Error creating new image")
    }

    // Encode new image
    err = imageHandler.Encode(newFileImg, newImage)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", errors.New("Error encoding response")
    }
    newFileImg.Close()

    err = imageHandler.Convert(newImageTempPath, imageParameters.Quality)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", err
    }

    err = os.Rename(newImageTempPath, newImageRealPath)
    if err != nil {
        os.Remove(newImageTempPath)
        return "", "", errors.New("Error moving file")
    }

    return newImageRealPath, responseType, nil

}

func BuildResponse(w http.ResponseWriter, imagePath string, contentType string) error {
    img, err := os.Open(imagePath)
    if err != nil {
        return errors.New("Error reading from optimized file")
    }
    defer img.Close()
    w.Header().Set("Content-Type", contentType) // <-- set the content-type header
    io.Copy(w, img)
    return nil
}
