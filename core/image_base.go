package core

import (
	"errors"
	"image"
	"io"
	"os"
)

type ImageHandler interface {
	ImageType() string
	ImageExtension() string
	Decode(reader io.Reader) (image.Image, error)
	Encode(newImgFile *os.File, newImage image.Image) error
	Optimize(newImageTempPath string, quality uint) error
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

func extensionToImageType(extension string) (string, error) {
	switch extension {
	case "jpeg":
		return "image/jpeg", nil
	case "jpg":
		return "image/jpeg", nil
	case "png":
		return "image/png", nil
	default:
		return "", errors.New("Unsupported Extension")
	}
}
