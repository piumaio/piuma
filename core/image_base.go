package core

import (
	"errors"
	"image"
	"io"
	"os"
)

type ImageHandler interface {
	ImageType() string
	Decode(reader io.Reader) (image.Image, error)
	Encode(newImgFile *os.File, newImage image.Image) error
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
