package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"

	"github.com/nfnt/resize"
)

// Optimize performs the heavy lifting of image transformation:
//  1. Instantiates an ImageHandler based on original Content-Type
//  2. Decodes the temporary file into an in-memory image.Image
//  3. Bounds requested width/height to original dimensions (0 preserves axis)
//  4. Resizes (NearestNeighbor for speed)
//  5. Optionally switches handler if Convert directive supplied
//  6. Encodes using fixed quality or adaptive DSSIM search
//  7. Penalizes format if resulting file larger than original by removing it
//     from future auto-negotiation choices
//  8. Persists optimized bytes to PathMedia and returns its path + mime
//
// Errors at any step clean up temp file and are propagated.
func Optimize(response *http.Response, imageParameters *ImageParameters, options *Options) (string, string, error) {
	responseType := response.Header.Get("Content-Type")

	imageHandler, err := NewImageHandler(responseType)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}

	file, err := os.Open(options.PathTemp)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", err
	}

	img, err := imageHandler.Decode(file)
	if err != nil {
		os.Remove(options.PathTemp)
		return "", "", errors.New("error decoding response")
	}

	imageParameters.Width = uint(math.Min(float64(imageParameters.Width), float64(img.Bounds().Max.X)))
	imageParameters.Height = uint(math.Min(float64(imageParameters.Height), float64(img.Bounds().Max.Y)))

	newImage := resize.Resize(imageParameters.Width, imageParameters.Height, img, resize.NearestNeighbor)

	// Convert: pick another encoder if requested and not default (default means
	// keep original format). "auto" already resolved before Optimize.
	if imageParameters.Convert != "" && imageParameters.Convert != "default" {
		imageHandler, err = NewImageHandlerByExtension(imageParameters.Convert)
		if err != nil {
			os.Remove(options.PathTemp)
			return "", "", err
		}
	}

	var newFileBuffer bytes.Buffer
	// Adaptive: run perceptual compression search except for AVIF where quality
	// mapping already nonlinear and external tool control differs.
	if imageParameters.AdaptiveQuality && imageHandler.ImageType() != "avif" {
		err = CompressByDSSIM(&newImage, &newFileBuffer, &imageHandler, math.Abs(float64(imageParameters.Quality)-100)/10000)
	} else {
		err = imageHandler.Encode(&newFileBuffer, newImage, imageParameters.Quality)
	}

	defer os.Remove(options.PathTemp)
	if err != nil {
		return "", "", err
	}

	// If optimized file is paradoxically bigger, we log and demote the chosen
	// format within auto-conf (unless original already same type) to encourage
	// alternative future selections.
	if fileStat.Size() < int64(newFileBuffer.Len()) {
		log.Printf("[%s] [%s] Elaborated image is bigger than original...\n", imageParameters.GetUrlString(), response.Request.URL)

		imageParameters.Convert = "auto"
		autoConfPath := path.Join(path.Dir(options.PathMedia), imageParameters.GenerateHash(response))
		if responseType != imageHandler.ImageType() {
			RemoveImageHandlerFromAutoConf(autoConfPath, imageHandler.ImageType())
		}
	}

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
}
