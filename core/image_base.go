package core

import (
	"bytes"
	"encoding/gob"
	"errors"
	"image"
	"io"
	"net/http"
	"os"

	"github.com/elnormous/contenttype"
)

var imageHandlers = map[string]ImageHandler{
	"image/jpeg": &JPEGHandler{},
	"image/png":  &PNGHandler{},
	"image/webp": &WebPHandler{},
	"image/avif": &AvifHandler{},
}

var imageHandlersbyExtension = map[string]ImageHandler{
	"jpeg":          &JPEGHandler{},
	"jpg":           &JPEGHandler{},
	"png":           &PNGHandler{},
	"webp_lossless": &WebPLosslessHandler{},
	"webp":          &WebPHandler{},
	"avif":          &AvifHandler{},
}

type ImageHandler interface {
	ImageType() string
	ImageExtension() string
	SupportsTransparency() bool
	Decode(reader io.Reader) (image.Image, error)
	Encode(newImgFile io.Writer, newImage image.Image, quality uint) error
}

func NewImageHandler(imageType string) (ImageHandler, error) {
	if handler, ok := imageHandlers[imageType]; ok {
		return handler, nil
	}
	return nil, errors.New("Unsupported Image type")
}

func NewImageHandlerByExtension(extension string) (ImageHandler, error) {
	if handler, ok := imageHandlersbyExtension[extension]; ok {
		return handler, nil
	}
	return nil, errors.New("Unsupported Extension")
}

func NewImageHandlerByBytes(buffer io.Reader) (ImageHandler, error) {
	firstBytes := make([]byte, 512)
	_, err := buffer.Read(firstBytes)
	if err != nil {
		return nil, errors.New("Unsupported Extension")
	}
	contentType := http.DetectContentType(firstBytes)

	if contentType == "application/octet-stream" {
		if bytes.Compare(firstBytes[8:12], []byte("avif")) == 0 {
			return &AvifHandler{}, nil
		}
		return nil, errors.New("Unsupported Extension")
	} else {
		return NewImageHandler(contentType)
	}
}

func AutoImageHandler(clientRequest *http.Request, imageResponse *http.Response, autoConfPath string) (ImageHandler, error) {
	imageHandler, err := NewImageHandler(imageResponse.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	var availableMediaTypes []contenttype.MediaType

	if file, err := os.Open(autoConfPath); err == nil {
		dec := gob.NewDecoder(file)
		dec.Decode(&availableMediaTypes)
		file.Close()
	} else {
		availableMediaTypes = []contenttype.MediaType{
			contenttype.NewMediaType("image/png"),
			contenttype.NewMediaType("image/webp"),
			contenttype.NewMediaType("image/avif"),
		}
		if !imageHandler.SupportsTransparency() {
			availableMediaTypes = append([]contenttype.MediaType{contenttype.NewMediaType("image/jpeg")}, availableMediaTypes...)
		}

		if file, err := os.Create(autoConfPath); err == nil {
			enc := gob.NewEncoder(file)
			enc.Encode(availableMediaTypes)
			file.Close()
		}
	}

	accepted, _, err := contenttype.GetAcceptableMediaType(clientRequest, availableMediaTypes)
	if err != nil {
		return nil, errors.New("Error while trying to parse Accept header")
	}
	imageHandler, err = NewImageHandler(accepted.String())
	if err != nil {
		return nil, err
	}

	return imageHandler, nil
}

func RemoveImageHandlerFromAutoConf(autoConfPath string, imageType string) error {
	var availableMediaTypes []contenttype.MediaType
	var err error

	if file, err := os.Open(autoConfPath); err == nil {
		dec := gob.NewDecoder(file)
		dec.Decode(&availableMediaTypes)

		temp := availableMediaTypes[:0]
		for _, x := range availableMediaTypes {
			if x.String() != imageType {
				temp = append(temp, x)
			}
		}
		availableMediaTypes = temp
		file.Close()
	} else {
		return err
	}

	if file, err := os.Create(autoConfPath); err == nil {
		enc := gob.NewEncoder(file)
		enc.Encode(availableMediaTypes)
		file.Close()
	}
	return err
}

func GetAllImageHandlers() map[string]ImageHandler {
	return imageHandlers
}

func GetAllImageHandlersByExtension() map[string]ImageHandler {
	return imageHandlersbyExtension
}
