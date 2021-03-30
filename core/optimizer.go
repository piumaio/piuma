package core

import (
	"errors"
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

	if imageParameters.Convert != "" {
		imageHandler, err = NewImageHandlerByExtension(imageParameters.Convert)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", errors.New("Error while converting image handler")
		}
	}

	finalFile, err := os.Create(options.PathTemp)
	defer file.Close()
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}

	_, isFast := imageHandler.(FastImageHandler)
	if isFast {
		imageHandler.(FastImageHandler).Encode(finalFile, newImage, imageParameters.Quality)
	} else {
		advImageHandler := imageHandler.(AdvancedImageHandler)

		err = advImageHandler.Encode(finalFile, newImage)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", errors.New("Error encoding response")
		}
		finalFile.Close()

		err = advImageHandler.Optimize(options.PathTemp, imageParameters.Quality)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", err
		}
	}

	err = os.Rename(options.PathTemp, options.PathMedia)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", errors.New("Error moving file")
	}

	return options.PathMedia, imageHandler.ImageType(), nil
}
