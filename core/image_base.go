package core

type ImageHandler interface {
	ImageType() string
}

func NewImageHandler(imageType string) (ImageHandler, error) {
	switch generatorType {
	case "image/jpeg":
		return &JPEGHandler{}
	case "image/png":
		return &PNGHandler{}
	default:
		return nil, errors.New("Unsupported Image type")
	}
}
