package core

import (
	"image"
	"image/jpeg"
	"io"
	"os"
)

type JPEGHandler struct {
}

func (j *JPEGHandler) ImageType() string {
	return "image/png"
}

func (j *JPEGHandler) Decode(reader io.Reader) (image.Image, error) {
	return jpeg.Decode(reader)
}

func (j *JPEGHandler) Encode(newImgFile *os.File, newImage image.Image) error {
	return jpeg.Encode(imgFile, image, nil)
}
