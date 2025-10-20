package core

import (
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
)

// PNGHandler implements ImageHandler using the stdlib PNG encoder plus a pass
// through the external optipng binary for further lossless compression.
type PNGHandler struct {
	ImageHandler
}

func (p *PNGHandler) ImageType() string {
	return "image/png"
}

func (p *PNGHandler) ImageExtension() string {
	return "png"
}

func (p *PNGHandler) SupportsTransparency() bool {
	return true
}

func (p *PNGHandler) Decode(reader io.Reader) (image.Image, error) {
	return png.Decode(reader)
}

// Encode writes a compressed PNG using optipng for optimization. Quality is
// ignored (PNG is lossless); the parameter is accepted for interface parity.
func (p *PNGHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	file, err := os.CreateTemp("", "png_image")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	err = png.Encode(file, newImage)
	if err != nil {
		return err
	}
	file.Close()

	args := []string{file.Name()}
	cmd := exec.Command("optipng", args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("optipng command not working")
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
