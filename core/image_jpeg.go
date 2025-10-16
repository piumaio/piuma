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

// seam for external jpegoptim invocation; overridden in tests to simulate success/failure without tool.
var jpegoptimCmd = func(args ...string) *exec.Cmd { return exec.Command("jpegoptim", args...) }

// JPEGHandler implements lossy JPEG encoding. After initial encode it runs
// jpegoptim to enforce size/quality constraints (progressive, strip metadata).
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

// Encode writes a JPEG with the requested quality (0-100) then invokes
// jpegoptim to perform further compression. Returns error if jpegoptim fails.
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
	cmd := jpegoptimCmd(args...)
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
