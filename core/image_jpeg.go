package core

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
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
	return jpeg.Encode(newImgFile, newImage, nil)
}

func (j *JPEGHandler) Convert(newImageTempPath string, quality uint) error {
	args := []string{fmt.Sprintf("--max=%d", quality), newImageTempPath}
	cmd := exec.Command("jpegoptim", args...)
	err := cmd.Run()
	if err != nil {
		return errors.New("Jpegoptim command not working")
	}

	return nil
}
