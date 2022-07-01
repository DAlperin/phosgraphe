package transforms

import (
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strconv"
	"strings"
)

type rotateTransformation struct{}

func (rT rotateTransformation) Handles(transform string) (bool, error) {
	accepted := []string{"rotate_\\w+", "r_\\w+"}
	return genericHandles(accepted, transform)
}

func (rT rotateTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImageBlob(file)
	if err != nil {
		return nil, err
	}

	angle, err := strconv.Atoi(strings.Split(instruction, "_")[1])
	if err != nil {
		return nil, err
	}

	pixel := imagick.NewPixelWand()
	defer pixel.Destroy()

	if mw.GetFormat() == "png" {
		pixel.SetColor("transparent")
	}

	err = mw.RotateImage(pixel, float64(angle))
	if err != nil {
		return nil, err
	}

	res := mw.GetImageBlob()
	return res, nil
}
