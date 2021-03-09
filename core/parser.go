package core

import (
	"fmt"
	"strconv"
	"strings"
)

// ImageParameters represents the parameters for optimization
type ImageParameters struct {
	Width   uint
	Height  uint
	Quality uint
	Convert string
}

func (imParams *ImageParameters) prepareHashData() string {
	return fmt.Sprint(imParams.Width, imParams.Height, imParams.Quality, imParams.Convert)
}

// Parser extracts width, height and quality from the provided parameters.
func Parser(name string) (ImageParameters, error) {
	stringSlice := strings.Split(name, "/")
	dimqual := strings.Split(stringSlice[0], ":")

	dimQualityArray := strings.Split(dimqual[0], "_")
	arrayOfInt := getDefaultParameters()

	var err error
	var tmpr int

	for i := 0; i < len(dimQualityArray); i++ {
		tmpr, err = strconv.Atoi(dimQualityArray[i])
		if err != nil {
			return ImageParameters{}, err
		}
		arrayOfInt[i] = uint(tmpr)
	}

	var convertTo string = ""
	if len(dimqual) > 1 {
		convertTo = dimqual[1]
	}

	parameters := ImageParameters{
		Width:   arrayOfInt[0],
		Height:  arrayOfInt[1],
		Quality: arrayOfInt[2],
		Convert: convertTo,
	}
	return parameters, nil
}

// getDefaultParameters creates an the default parameters
// for optimization
func getDefaultParameters() []uint {
	defaultParams := make([]uint, 3)
	defaultParams[0] = 0
	defaultParams[1] = 0
	defaultParams[2] = 100

	return defaultParams
}
