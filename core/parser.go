package core

import (
	"fmt"
	"strconv"
	"strings"
)

// ImageParameters represents the parameters for optimization
type ImageParameters struct {
	Width           uint
	Height          uint
	Quality         uint
	AdaptiveQuality bool
	Convert         string
}

func (imParams *ImageParameters) PrepareHashData() string {
	return fmt.Sprint(imParams.Width, imParams.Height, imParams.Quality, imParams.AdaptiveQuality, imParams.Convert)
}

func (imParams *ImageParameters) GetUrlString() string {
	urlString := fmt.Sprintf("%d_%d_%d", imParams.Width, imParams.Height, imParams.Quality)
	if imParams.AdaptiveQuality {
		urlString += "a"
	}
	if imParams.Convert != "" {
		urlString += fmt.Sprintf(":%s", imParams.Convert)
	}
	return urlString
}

// Parser extracts width, height and quality from the provided parameters.
func Parser(name string) (ImageParameters, error) {
	stringSlice := strings.Split(name, "/")
	dimqual := strings.Split(stringSlice[0], ":")

	dimQualityArray := strings.Split(dimqual[0], "_")
	arrayOfInt := getDefaultParameters()

	var err error
	var tmpr int
	isQualityAdaptive := false

	for i := 0; i < len(dimQualityArray); i++ {
		data := dimQualityArray[i]
		if i == 2 && strings.HasSuffix(data, "a") {
			lenData := len(data)
			data = data[:lenData-1]
			isQualityAdaptive = true
		}
		tmpr, err = strconv.Atoi(data)
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
		Width:           arrayOfInt[0],
		Height:          arrayOfInt[1],
		Quality:         arrayOfInt[2],
		AdaptiveQuality: isQualityAdaptive,
		Convert:         convertTo,
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
