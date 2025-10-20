package core

import (
	"bytes"
	"encoding/gob"
	"errors"
	"image"
	"io"
	"net/http"
	"os"

	"github.com/elnormous/contenttype"
)

// imageHandlers maps MIME types to their concrete handlers for decoding and
// encoding. Used for Content-Type based resolution.
var imageHandlers = map[string]ImageHandler{
	"image/jpeg": &JPEGHandler{},
	"image/png":  &PNGHandler{},
	"image/webp": &WebPHandler{},
	"image/avif": &AvifHandler{},
}

// imageHandlersbyExtension maps short extension tokens (parse Convert input)
// to handlers. Includes distinct lossless WebP option.
var imageHandlersbyExtension = map[string]ImageHandler{
	"jpeg":          &JPEGHandler{},
	"jpg":           &JPEGHandler{},
	"png":           &PNGHandler{},
	"webp_lossless": &WebPLosslessHandler{},
	"webp":          &WebPHandler{},
	"avif":          &AvifHandler{},
}

// ImageHandler abstracts format-specific operations. Implementations should:
//
//	ImageType()          -> return canonical MIME type
//	ImageExtension()     -> short token used in Convert directives
//	SupportsTransparency() -> whether alpha channel can be preserved
//	Decode(io.Reader)    -> produce image.Image from encoded bytes
//	Encode(io.Writer, image.Image, quality) -> write encoded bytes
//
// Quality meaning depends on encoder (lossless handlers may ignore it).
type ImageHandler interface {
	ImageType() string
	ImageExtension() string
	SupportsTransparency() bool
	Decode(reader io.Reader) (image.Image, error)
	Encode(newImgFile io.Writer, newImage image.Image, quality uint) error
}

// NewImageHandler returns a handler based on MIME Content-Type string.
// Returns error if unsupported.
func NewImageHandler(imageType string) (ImageHandler, error) {
	if handler, ok := imageHandlers[imageType]; ok {
		return handler, nil
	}
	return nil, errors.New("unsupported image type")
}

// NewImageHandlerByExtension returns a handler mapped from Convert extension
// token (e.g. "webp", "avif", "webp_lossless"). Returns error if unsupported.
func NewImageHandlerByExtension(extension string) (ImageHandler, error) {
	if handler, ok := imageHandlersbyExtension[extension]; ok {
		return handler, nil
	}
	return nil, errors.New("unsupported extension")
}

// NewImageHandlerByBytes inspects the first 512 bytes of a file and attempts
// to detect its Content-Type. Special-cases AVIF which may appear as
// application/octet-stream and inspects magic bytes. Returns error if unknown.
func NewImageHandlerByBytes(buffer io.Reader) (ImageHandler, error) {
	firstBytes := make([]byte, 512)
	_, err := buffer.Read(firstBytes)
	if err != nil {
		return nil, errors.New("unsupported extension")
	}
	contentType := http.DetectContentType(firstBytes)

	if contentType == "application/octet-stream" {
		if bytes.Equal(firstBytes[8:12], []byte("avif")) {
			return &AvifHandler{}, nil
		}
		return nil, errors.New("unsupported extension")
	} else {
		return NewImageHandler(contentType)
	}
}

// AutoImageHandler performs dynamic format negotiation based on the client's
// Accept header and a persisted set of available media types on disk (gob
// file). The persisted list allows pruning of poor-performing formats over
// time. preferredConverts optionally restricts candidate formats (result of
// "auto:..." directive). Returns the chosen handler or error.
func AutoImageHandler(clientRequest *http.Request, imageResponse *http.Response, autoConfPath string, preferredConverts []string) (ImageHandler, error) {
	imageHandler, err := NewImageHandler(imageResponse.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	var availableMediaTypes []contenttype.MediaType

	if file, err := os.Open(autoConfPath); err == nil {
		dec := gob.NewDecoder(file)
		dec.Decode(&availableMediaTypes)
		file.Close()
	} else {
		availableMediaTypes = []contenttype.MediaType{
			contenttype.NewMediaType("image/png"),
			contenttype.NewMediaType("image/webp"),
			contenttype.NewMediaType("image/avif"),
		}
		if !imageHandler.SupportsTransparency() {
			availableMediaTypes = append([]contenttype.MediaType{contenttype.NewMediaType("image/jpeg")}, availableMediaTypes...)
		}

		if file, err := os.Create(autoConfPath); err == nil {
			enc := gob.NewEncoder(file)
			enc.Encode(availableMediaTypes)
			file.Close()
		}
	}

	if len(preferredConverts) > 0 {
		allowedConverts := make(map[string]bool)
		allowedMediaTypes := []contenttype.MediaType{}

		for _, c := range preferredConverts {
			h, err := NewImageHandlerByExtension(c)
			if err == nil {
				allowedConverts[h.ImageType()] = true
			}
		}

		for _, mt := range availableMediaTypes {
			if _, ok := allowedConverts[mt.String()]; ok {
				allowedMediaTypes = append(allowedMediaTypes, mt)
			}
		}
		availableMediaTypes = allowedMediaTypes
	}

	accepted, _, err := contenttype.GetAcceptableMediaType(clientRequest, availableMediaTypes)
	if err != nil {
		return nil, errors.New("error while trying to parse accept header")
	}
	imageHandler, err = NewImageHandler(accepted.String())
	if err != nil {
		return nil, err
	}

	return imageHandler, nil
}

// RemoveImageHandlerFromAutoConf rewrites the gob file at autoConfPath
// removing the specified imageType from future auto negotiation.
func RemoveImageHandlerFromAutoConf(autoConfPath string, imageType string) error {
	var availableMediaTypes []contenttype.MediaType
	var err error

	if file, err := os.Open(autoConfPath); err == nil {
		dec := gob.NewDecoder(file)
		dec.Decode(&availableMediaTypes)

		temp := availableMediaTypes[:0]
		for _, x := range availableMediaTypes {
			if x.String() != imageType {
				temp = append(temp, x)
			}
		}
		availableMediaTypes = temp
		file.Close()
	} else {
		return err
	}

	if file, err := os.Create(autoConfPath); err == nil {
		enc := gob.NewEncoder(file)
		enc.Encode(availableMediaTypes)
		file.Close()
	}
	return err
}

// GetAllImageHandlers exposes internal MIME map (read-only usage intended).
func GetAllImageHandlers() map[string]ImageHandler {
	return imageHandlers
}

// GetAllImageHandlersByExtension returns the extension-based handler map.
func GetAllImageHandlersByExtension() map[string]ImageHandler {
	return imageHandlersbyExtension
}
