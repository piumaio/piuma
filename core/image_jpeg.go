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
	defer os.Remove(file.Name())

	err = jpeg.Encode(file, newImage, &jpeg.Options{Quality: int(quality)})
	if err != nil {
		return err
	}
	file.Close()

	args := []string{fmt.Sprintf("--max=%d", quality), "--all-progressive", "-s", "-o", file.Name()}
	cmd := exec.Command("jpegoptim", args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("Jpegoptim command not working")
	}

	file, err = os.Open(file.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(newImgFile, file)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}
