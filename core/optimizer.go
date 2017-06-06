package core

import (
    "net/http"
    "io"
    "os"
    "os/exec"
    "crypto/sha1"
    "encoding/base64"
    "image"
    "image/jpeg"
    "image/png"
    "github.com/nfnt/resize"
    "errors"
    "sync"
    "path/filepath"
    "fmt"
)


func Optimize(originalUrl string, width uint, height uint, quality uint, pathtemp string, pathmedia string) (string, string, error) {

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
    hash.Write([]byte(fmt.Sprint(width, height, quality, originalUrl, responseType, size, lastModified)))
    newFileName := base64.URLEncoding.EncodeToString(hash.Sum(nil))

    newImageTempPath := filepath.Join(pathtemp, newFileName)
    newImageRealPath := filepath.Join(pathmedia, newFileName)

    // Check if file exists
    if _, err := os.Stat(newImageRealPath); err == nil {
      return newImageRealPath, responseType, nil
    }

    // Decode and resize
    var reader io.Reader = response.Body
    var newFileImg *os.File = nil
    var mu = &sync.Mutex{}

    mu.Lock()
    if _, err := os.Stat(newImageTempPath); err == nil {
        return "", "", errors.New("Still elaborating")
    } else {
        newFileImg, err = os.Create(newImageTempPath)
    }
    mu.Unlock()

    var img image.Image

    if responseType == "image/jpeg" {
        img, err = jpeg.Decode(reader)
    } else if responseType == "image/png" {
        img, err = png.Decode(reader)
    } else {
        return "", "", errors.New("Format not supported")
    }

    if err != nil {
        return "", "", errors.New("Error decoding response")
    }

    newImage := resize.Resize(width, height, img, resize.NearestNeighbor)

    if err != nil {
        return "", "", errors.New("Error creating new image")
    }

    // Encode new image
    if responseType == "image/jpeg" {
        err = jpeg.Encode(newFileImg, newImage, nil)
        if err != nil {
            return "", "", errors.New("Error encoding response")
        }
    } else if responseType == "image/png" {
        err = png.Encode(newFileImg, newImage)
        if err != nil {
            return "", "", errors.New("Error encoding response")
        }
    }
    newFileImg.Close()

    if responseType == "image/jpeg" {
        args := []string{fmt.Sprintf("--max=%d", quality), newImageTempPath}
        cmd := exec.Command("jpegoptim", args...)
        err := cmd.Run()
        if err != nil {
            return "", "", errors.New("Jpegoptim command not working")
        }
    }else if responseType == "image/png" {
        var qualityMin = quality-10
        args := []string{fmt.Sprintf("--quality=%[1]d-%[2]d", qualityMin, quality), newImageTempPath, "-f", "--ext=\"\""}
        fmt.Println(args)
        cmd := exec.Command("pngquant", args...)
        err := cmd.Run()
        if err != nil {
            fmt.Println(err)
            return "", "", errors.New("Pngquant command not working")
        }
    }

    err = os.Rename(newImageTempPath, newImageRealPath)
    if err != nil {
        return "", "", errors.New("Error moving file")
    }

    return newImageRealPath, responseType, nil
}


func BuildResponse (w http.ResponseWriter, imagePath string, contentType string) (error){
    img, err := os.Open(imagePath)
    if err != nil {
        return errors.New("Error reading from optimized file")
    }
    defer img.Close()
    w.Header().Set("Content-Type", contentType) // <-- set the content-type header
    io.Copy(w, img)
    return nil
}
