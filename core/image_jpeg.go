package core

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type JPEGHandler struct {
	ImageHandler
}

func (j *JPEGHandler) ImageType() string {
	return "image/jpeg"
}

func (j *JPEGHandler) ImageExtension() string {
	return "jpg"
}

func (j *JPEGHandler) SupportsTransparency() bool {
	return false
}

func (j *JPEGHandler) Decode(reader io.Reader) (image.Image, error) {
	return jpeg.Decode(reader)
}

func (j *JPEGHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	file, err := ioutil.TempFile("", "jpeg_image")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	err = jpeg.Encode(file, newImage, nil)
	if err != nil {
		return err
	}

	args := []string{fmt.Sprintf("--max=%d", quality), file.Name()}
	cmd := exec.Command("jpegoptim", args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("Jpegoptim command not working")
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = io.Copy(newImgFile, file)
	if err != nil {
		return err
	}

	return nil
}
