package core

import (
	"image"
	"io"
)

type PNGHandler struct {
}

func (png *PNGHandler) ImageType() string {
	return "image/png"
}

func (png *PNGHandler) Decode(reader io.Reader) (image.Image, error) {
	return png.Decode(reader)
}
