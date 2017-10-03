package core

import (
	"errors"
	"image"
	"io"
)

type ImageHandler interface {
	ImageType() string
	Decode(io.Reader) (image.Image, error)
}

func NewImageHandler(imageType string) (ImageHandler, error) {
	switch imageType {
	case "image/jpeg":
		return &JPEGHandler{}, nil
	case "image/png":
		return &PNGHandler{}, nil
	default:
		return nil, errors.New("Unsupported Image type")
	}
}
