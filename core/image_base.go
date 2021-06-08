package core

import (
	"bytes"
	"errors"
	"image"
	"io"
	"net/http"

	"github.com/elnormous/contenttype"
)

type ImageHandler interface {
	ImageType() string
	ImageExtension() string
	SupportsTransparency() bool
	Decode(reader io.Reader) (image.Image, error)
	Encode(newImgFile io.Writer, newImage image.Image, quality uint) error
}

func NewImageHandler(imageType string) (ImageHandler, error) {
	switch imageType {
	case "image/jpeg":
		return &JPEGHandler{}, nil
	case "image/png":
		return &PNGHandler{}, nil
	case "image/webp":
		return &WebPHandler{}, nil
	case "image/avif":
		return &AvifHandler{}, nil
	default:
		return nil, errors.New("Unsupported Image type")
	}
}

func NewImageHandlerByExtension(extension string) (ImageHandler, error) {
	switch extension {
	case "jpeg":
		return &JPEGHandler{}, nil
	case "jpg":
		return &JPEGHandler{}, nil
	case "png":
		return &PNGHandler{}, nil
	case "webp_lossless":
		return &WebPLosslessHandler{}, nil
	case "webp":
		return &WebPHandler{}, nil
	case "avif":
		return &AvifHandler{}, nil
	default:
		return nil, errors.New("Unsupported Extension")
	}
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

func AutoImageHandler(clientRequest *http.Request, imageResponse *http.Response) (ImageHandler, error) {
	imageHandler, err := NewImageHandler(imageResponse.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	availableMediaTypes := []contenttype.MediaType{
		contenttype.NewMediaType("image/png"),
		contenttype.NewMediaType("image/webp"),
		contenttype.NewMediaType("image/avif"),
	}
	if !imageHandler.SupportsTransparency() {
		availableMediaTypes = append([]contenttype.MediaType{contenttype.NewMediaType("image/jpeg")}, availableMediaTypes...)
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
