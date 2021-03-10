package core

import (
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
)

type PNGHandler struct {
	AdvancedImageHandler
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

func (p *PNGHandler) Encode(newImgFile *os.File, newImage image.Image) error {
	return png.Encode(newImgFile, newImage)
}

func (p *PNGHandler) Optimize(newImageTempPath string, quality uint) error {
	var err error
	var cmd *exec.Cmd

	default_args := []string{newImageTempPath}

	cmd = exec.Command("optipng", default_args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("OptiPNG command not working")
	}

	return nil
}
