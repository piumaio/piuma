package core

import (
	"image"
	"io"

	"github.com/chai2010/webp"
)

type WebPHandler struct {
	ImageHandler
}

func (w *WebPHandler) ImageType() string {
	return "image/webp"
}

func (w *WebPHandler) ImageExtension() string {
	return "webp"
}

func (w *WebPHandler) Decode(reader io.Reader) (image.Image, error) {
	return webp.Decode(reader)
}

func (w *WebPHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	return webp.Encode(newImgFile, newImage, &webp.Options{Lossless: false, Quality: float32(quality)})
}

type WebPLosslessHandler struct {
	WebPHandler
}

func (w *WebPLosslessHandler) ImageExtension() string {
	return "webp_lossless"
}

func (w *WebPLosslessHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	return webp.Encode(newImgFile, newImage, &webp.Options{Lossless: true, Quality: float32(quality)})
}
