package core

import (
	"bytes"
	"errors"
	"image"
	"io"
	"net/http"
)

type ImageHandler interface {
	ImageType() string
	ImageExtension() string
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
