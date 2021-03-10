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
}

type AdvancedImageHandler interface {
	ImageHandler
	Encode(newImgFile *os.File, newImage image.Image) error
	Optimize(newImageTempPath string, quality uint) error
}

type FastImageHandler interface {
	ImageHandler
	Encode(newImgFile *os.File, newImage image.Image, quality uint) error
}

func NewImageHandler(imageType string) (ImageHandler, error) {
	switch imageType {
	case "image/jpeg":
		return &JPEGHandler{}, nil
	case "image/png":
		return &PNGHandler{}, nil
	case "image/webp":
		return &WebPHandler{}, nil
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
	default:
		return nil, errors.New("Unsupported Extension")
	}
}
