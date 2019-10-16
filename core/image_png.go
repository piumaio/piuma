package core

import (
    "errors"
    "fmt"
    "image"
    "image/png"
    "io"
    "os"
    "os/exec"
)

type PNGHandler struct {
}

func (p *PNGHandler) ImageType() string {
    return "image/png"
}

func (p *PNGHandler) Decode(reader io.Reader) (image.Image, error) {
    return png.Decode(reader)
}

func (p *PNGHandler) Encode(newImgFile *os.File, newImage image.Image) error {
    return png.Encode(newImgFile, newImage)
}

func (p *PNGHandler) Convert(newImageTempPath string, quality uint) error {
    var err error
    var cmd *exec.Cmd

    default_args := []string{newImageTempPath, "-f", "--ext=.png", "--skip-if-larger", "--strip"}

    if quality != 100 {
        var qualityMin = quality - 10
        qualityParameter := fmt.Sprintf("--quality=%[1]d-%[2]d", qualityMin, quality)
        args := append([]string{qualityParameter}, default_args...)
        cmd = exec.Command("pngquant", args...)
        err = cmd.Run()
        if err == nil {
            return nil
        }
    }
    cmd = exec.Command("pngquant", default_args...)
    err = cmd.Run()
    if err != nil {
        return errors.New("Pngquant command not working")
    }

    return nil
}
