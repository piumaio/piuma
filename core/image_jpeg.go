package core

import (
	"image"
	"image/jpeg"
	"io"
)

type JPEGHandler struct {
}

func (png *JPEGHandler) ImageType() string {
	return "image/png"
}

func (png *JPEGHandler) Decode(reader io.Reader) (image.Image, error) {
	return jpeg.Decode(reader)
}
