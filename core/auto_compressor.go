package core

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// seam for dssim invocation
var dssimCmd = func(args ...string) *exec.Cmd { return exec.Command("dssim", args...) }

// MaxIterations bounds the adaptive quality binary search steps to guarantee
// predictable CPU usage. After MaxIterations we accept the last candidate.
const MaxIterations = 4

// CompressByDSSIM adaptively finds an encoding quality that keeps DSSIM value
// (structural dissimilarity) below threshold using a bounded binary search.
// threshold typically scales from requested quality (lower threshold -> better
// visual fidelity). The handler's Encode/Decode are used to round-trip bytes.
// Writes final encoded bytes to newImgFile.
func CompressByDSSIM(original *image.Image, newImgFile io.Writer, handler *ImageHandler, threshold float64) error {
	startQuality := 0
	endQuality := 100
	iterations := 0

	originalFile, err := createTempPNG(original)
	if err != nil {
		return errors.New("Cannot create temp images for dssim")
	}
	defer os.Remove(originalFile.Name())
	defer originalFile.Close()

	buf := new(bytes.Buffer)

	currentQuality := startQuality + int(math.Abs(float64(startQuality-endQuality))/2)
	err = (*handler).Encode(buf, *original, uint(currentQuality))
	if err != nil {
		return err
	}
	for iterations < MaxIterations {
		imageCompressed, err := (*handler).Decode(buf)
		if err != nil {
			return err
		}

		dssimValue, err := getDSSIMValue(originalFile, &imageCompressed)
		if err != nil {
			return err
		}

		if dssimValue < threshold {
			endQuality = int(currentQuality)
		} else {
			startQuality = int(currentQuality)
		}

		currentQuality = startQuality + int(math.Abs(float64(startQuality-endQuality))/2)
		err = (*handler).Encode(buf, *original, uint(currentQuality))
		if err != nil {
			return err
		}
		iterations++
	}

	_, err = io.Copy(newImgFile, buf)
	if err != nil {
		return err
	}

	return nil
}

// getDSSIMValue produces a DSSIM score invoking the external `dssim` tool on
// two temporary PNG files: original (file1) and the candidate image (image2).
// Returns parsed float or error if tool invocation fails.
func getDSSIMValue(file1 *os.File, image2 *image.Image) (float64, error) {
	file2, err := createTempPNG(image2)
	if err != nil {
		return -1, errors.New("Cannot create temp images for dssim")
	}
	defer file2.Close()
	defer os.Remove(file2.Name())

	args := []string{file1.Name(), file2.Name()}
	dssimValue, err := dssimCmd(args...).Output()
	if err != nil {
		return -1, errors.New("dssim command not working")
	}

	return strconv.ParseFloat(strings.Split(string(dssimValue), "\t")[0], 64)
}

// createTempPNG encodes an image to a temporary PNG file for external tool
// analysis. Caller must remove the file. Returns file handle positioned at end
// (decoder tools reopen). Used by DSSIM evaluation.
func createTempPNG(image *image.Image) (*os.File, error) {
	file, err := os.CreateTemp("", "dssim_image")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = png.Encode(file, *image)
	if err != nil {
		defer os.Remove(file.Name())
		return nil, err
	}

	return file, nil
}
