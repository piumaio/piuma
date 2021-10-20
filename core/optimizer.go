package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

func Optimize(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	responseType := response.Header.Get("Content-Type")

	imageHandler, err := NewImageHandler(responseType)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}

	file, err := os.Open(options.PathTemp)
	defer file.Close()
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}
	fileStat, err := file.Stat()
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}

	img, err := imageHandler.Decode(file)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", errors.New("Error decoding response")
	}

	newImage := resize.Resize(imageParameters.Width, imageParameters.Height, img, resize.NearestNeighbor)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", errors.New("Error creating new image")
	}

	if imageParameters.Convert != "" && imageParameters.Convert != "default" {
		imageHandler, err = NewImageHandlerByExtension(imageParameters.Convert)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", err
		}
	}

	var newFileBuffer bytes.Buffer
	if imageParameters.AdaptiveQuality {
		err = CompressByDSSIM(&newImage, &newFileBuffer, &imageHandler, math.Abs(float64(imageParameters.Quality)-100)/10000)
	} else {
		err = imageHandler.Encode(&newFileBuffer, newImage, imageParameters.Quality)
	}

	if fileStat.Size() > int64(newFileBuffer.Len()) {
		defer os.Remove(options.PathTemp)

		newFile, err := os.Create(options.PathMedia)
		if err != nil {
			return "", "", err
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, &newFileBuffer)
		if err != nil {
			return "", "", err
		}

		return options.PathMedia, imageHandler.ImageType(), nil
	} else {
		log.Printf("[%s] [%s] Elaborated image is bigger than original, going back to original...\n", imageParameters.GetUrlString(), response.Request.URL)

		err = os.Rename(options.PathTemp, options.PathMedia)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", err
		}

		imageHandler, err := NewImageHandler(responseType)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", err
		}

		return options.PathMedia, imageHandler.ImageType(), nil
	}
}
