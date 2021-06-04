package core

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const MaxIterations = 4

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

func getDSSIMValue(file1 *os.File, image2 *image.Image) (float64, error) {
	file2, err := createTempPNG(image2)
	if err != nil {
		return -1, errors.New("Cannot create temp images for dssim")
	}
	defer file2.Close()
	defer os.Remove(file2.Name())

	args := []string{file1.Name(), file2.Name()}
	dssimValue, err := exec.Command("dssim", args...).Output()
	if err != nil {
		return -1, errors.New("dssim command not working")
	}

	return strconv.ParseFloat(strings.Split(string(dssimValue), "\t")[0], 64)
}

func createTempPNG(image *image.Image) (*os.File, error) {
	file, err := ioutil.TempFile("", "dssim_image")
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
