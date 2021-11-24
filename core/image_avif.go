package core

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type AvifHandler struct {
	ImageHandler
}

func (a *AvifHandler) ImageType() string {
	return "image/avif"
}

func (a *AvifHandler) ImageExtension() string {
	return "avif"
}

func (a *AvifHandler) SupportsTransparency() bool {
	return true
}

func (a *AvifHandler) Decode(reader io.Reader) (image.Image, error) {
	avifFile, err := ioutil.TempFile("", "dec_image*.avif")
	if err != nil {
		return nil, err
	}
	defer avifFile.Close()
	defer os.Remove(avifFile.Name())

	_, err = io.Copy(avifFile, reader)
	if err != nil {
		return nil, err
	}

	pngFile, err := ioutil.TempFile("", "dec_image*.png")
	if err != nil {
		return nil, err
	}
	defer pngFile.Close()
	defer os.Remove(pngFile.Name())

	args := []string{"-q 100", avifFile.Name(), pngFile.Name()}
	cmd := exec.Command("avifdec", args...)
	err = cmd.Run()
	if err != nil {
		return nil, errors.New("avifdec command not working")
	}

	return png.Decode(pngFile)
}

func (a *AvifHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	pngFile, err := ioutil.TempFile("", "enc_image*.png")
	if err != nil {
		return err
	}
	defer pngFile.Close()
	defer os.Remove(pngFile.Name())

	err = png.Encode(pngFile, newImage)
	if err != nil {
		return err
	}

	avifFile, err := ioutil.TempFile("", "enc_image*.avif")
	if err != nil {
		return err
	}
	defer avifFile.Close()
	defer os.Remove(avifFile.Name())

	quality = (100 - quality) * 63 / 100

	args := []string{"--min", fmt.Sprint(quality), "--max", fmt.Sprint(quality), "--minalpha", fmt.Sprint(quality), "--maxalpha", fmt.Sprint(quality), pngFile.Name(), avifFile.Name()}
	cmd := exec.Command("avifenc", args...)
	err = cmd.Run()
	if err != nil {
		return errors.New("avifenc command not working")
	}

	_, err = io.Copy(newImgFile, avifFile)
	if err != nil {
		return err
	}

	return nil
}
