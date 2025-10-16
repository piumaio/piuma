package core

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ImageParameters holds all transformation directives parsed from the URL
// segment. Width and Height specify a desired bounding box (0 preserves
// original dimension for that axis). Quality is a percentage in [0,100].
// AdaptiveQuality toggles DSSIM-driven binary search for a perceptual quality
// (instead of fixed encoder quality). Convert optionally contains a target
// output format or the special forms:
//
//	"auto"                  -> negotiate best format from Accept header
//	"auto:<ext,...>"        -> same as auto but restricted to provided extensions
//
// For hashing we also consider original response metadata to guarantee cache
// invalidation when the source changes (e.g. Last-Modified, Content-Length).
type ImageParameters struct {
	Width           uint
	Height          uint
	Quality         uint
	AdaptiveQuality bool
	Convert         string
}

// GenerateHash produces a stable, collision-resistant identifier used for
// naming on-disk optimized images and auto-format configuration files. It
// combines transformation parameters and relevant original response headers
// (content type, size, last modified) plus the request URL so that any source
// change invalidates the cached optimized artifact. The hash is URL-safe
// base64 encoded for filesystem friendliness.
func (imParams *ImageParameters) GenerateHash(response *http.Response) string {
	responseType := response.Header.Get("Content-Type")
	size := response.Header.Get("Content-Length")
	lastModified := response.Header.Get("Last-Modified")

	hash := sha1.New()
	hash.Write([]byte(fmt.Sprint(imParams.Width, imParams.Height, imParams.Quality, imParams.AdaptiveQuality, imParams.Convert, response.Request.URL.String(), responseType, size, lastModified)))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

// GetUrlString reconstructs the canonical string representation of the
// parameters (width_height_quality["a"][:convert]). This mirrors the parsing
// format so it can be safely used for logging and debugging.
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

// Parser decodes the first path segment of the request into an ImageParameters
// value. Expected syntax:
//
//	<width>_<height>_<quality>[a][:convert][/...ignored]
//
// Where:
//
//	width,height,quality -> integers (0 height/width means "keep original")
//	optional trailing 'a' on quality -> enables AdaptiveQuality (DSSIM)
//	:convert -> either explicit extension (jpeg,png,webp,avif,webp_lossless)
//	            or auto / auto:<ext1,ext2,...>
//
// Remaining path components after the first '/' are ignored by the parser
// (handled by router for URL to fetch). Returns an error on malformed numbers.
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

		if convertTo == "auto" && len(dimqual) > 2 {
			convertTo = dimqual[1] + ":" + dimqual[2]
		}
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

// getDefaultParameters returns the zero-width/height and full-quality defaults
// used when components are omitted. Width=0/Height=0 implies no resizing,
// Quality=100 means keep best quality (subject to encoder semantics).
func getDefaultParameters() []uint {
	defaultParams := make([]uint, 3)
	defaultParams[0] = 0
	defaultParams[1] = 0
	defaultParams[2] = 100

	return defaultParams
}
