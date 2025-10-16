package core

import (
	"image"
	"io"

	"github.com/chai2010/webp"
)

// WebPHandler encodes/decodes standard (lossy) WebP images using library
// routines (no external binary required).
type WebPHandler struct {
	ImageHandler
}

func (w *WebPHandler) ImageType() string {
	return "image/webp"
}

func (w *WebPHandler) ImageExtension() string {
	return "webp"
}

func (w *WebPHandler) SupportsTransparency() bool {
	return true
}

func (w *WebPHandler) Decode(reader io.Reader) (image.Image, error) {
	return webp.Decode(reader)
}

// Encode writes a lossy WebP with given quality.
func (w *WebPHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	return webp.Encode(newImgFile, newImage, &webp.Options{Lossless: false, Quality: float32(quality)})
}

// WebPLosslessHandler encodes lossless WebP variants; quality is a hint.
type WebPLosslessHandler struct {
	WebPHandler
}

func (w *WebPLosslessHandler) ImageExtension() string {
	return "webp_lossless"
}

// Encode writes a lossless WebP regardless of requested quality.
func (w *WebPLosslessHandler) Encode(newImgFile io.Writer, newImage image.Image, quality uint) error {
	return webp.Encode(newImgFile, newImage, &webp.Options{Lossless: true, Quality: float32(quality)})
}
