package core

import (
	"image"
	"image/png"
	"io"
	"os"
)

type PNGHandler struct {
}

func (p *PNGHandler) ImageType() string {
	return "image/png"
}

func (p *PNGHandler) Decode(reader io.Reader) (image.Image, error) {
	return png.Decode(reader)
}

func (p *PNGHandler) Encode(newImgFile *os.File, newImage image.Image) error {
	return png.Encode(imgFile, image)
}
