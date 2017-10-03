package core

import "strings"
import "strconv"

// ImageParameters represents the parameters for optimization
type ImageParameters struct {
	Width   uint
	Height  uint
	Quality uint
}

// Parser extracts width, height and quality from the provided parameters.
func Parser(name string) (ImageParameters, error) {
	stringSlice := strings.Split(name, "/")
	var dimqual = stringSlice[0]

	dimQualityArray := strings.Split(dimqual, "_")
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

	parameters := ImageParameters{
		Width:   arrayOfInt[0],
		Height:  arrayOfInt[1],
		Quality: arrayOfInt[2],
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
