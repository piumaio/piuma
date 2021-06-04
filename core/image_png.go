package core

import (
	"errors"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type PNGHandler struct {
	ImageHandler
}

func (p *PNGHandler) ImageType() string {
	return "image/png"
}

func (p *PNGHandler) ImageExtension() string {
	return "png"
}

func (p *PNGHandler) Decode(reader io.Reader) (image.Image, error) {
	return png.Decode(reader)
}

func (p *PNGHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	file, err := ioutil.TempFile("", "png_image")
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	err = png.Encode(file, newImage)
	if err != nil {
		return err
	}

	args := []string{file.Name()}
	cmd := exec.Command("optipng", args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("OptiPNG command not working")
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
