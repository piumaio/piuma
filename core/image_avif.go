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

// seams for external avif tools
var avifencCmd = func(args ...string) *exec.Cmd { return exec.Command("avifenc", args...) }
var avifdecCmd = func(args ...string) *exec.Cmd { return exec.Command("avifdec", args...) }

// AvifHandler performs AVIF encoding/decoding via external CLI tools (avifenc,
// avifdec). PNG is used as an intermediate representation.
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

// Decode converts AVIF bytes into PNG using avifdec, then decodes PNG. Returns
// error if external tool invocation fails.
func (a *AvifHandler) Decode(reader io.Reader) (image.Image, error) {
	avifFile, err := ioutil.TempFile("", "dec_image*.avif")
	if err != nil {
		return nil, err
	}
	defer func() {
		avifFile.Close()
		os.Remove(avifFile.Name())
	}()

	if _, err = io.Copy(avifFile, reader); err != nil {
		return nil, err
	}

	pngFile, err := ioutil.TempFile("", "dec_image*.png")
	if err != nil {
		return nil, err
	}
	defer func() {
		pngFile.Close()
		os.Remove(pngFile.Name())
	}()

	// avifdec syntax: avifdec input.avif output.png
	args := []string{avifFile.Name(), pngFile.Name()}
	cmd := avifdecCmd(args...)
	if err = cmd.Run(); err != nil {
		return nil, errors.New("avifdec command not working")
	}

	// reset read pointer before decoding
	if _, err = pngFile.Seek(0, 0); err != nil {
		return nil, err
	}
	return png.Decode(pngFile)
}

// Encode writes AVIF by first encoding to PNG then invoking avifenc with a
// mapped quality value (AVIF quality scale differs: we derive min/max). Alpha
// quality matches luma quality.
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
	cmd := avifencCmd(args...)
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
