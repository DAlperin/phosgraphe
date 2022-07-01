package transforms

import (
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strconv"
	"strings"
)

type heightTransformation struct {
}

type widthTransformation struct {
}

func (eT heightTransformation) Handles(transform string) (bool, error) {
	accepted := []string{"height_\\d+", "^h_\\d+"}
	return genericHandles(accepted, transform)
}

func (eT heightTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImageBlob(file)
	if err != nil {
		return nil, err
	}

	width := mw.GetImageWidth()
	height, err := strconv.Atoi(strings.Split(instruction, "_")[1])

	err = mw.AdaptiveResizeImage(width, uint(height))
	if err != nil {
		return nil, err
	}

	res := mw.GetImageBlob()
	return res, nil
}

func (wT widthTransformation) Handles(transform string) (bool, error) {
	accepted := []string{"width_\\d+", "^w_\\d+"}
	return genericHandles(accepted, transform)
}

func (wT widthTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImageBlob(file)
	if err != nil {
		return nil, err
	}

	height := mw.GetImageHeight()

	width, err := strconv.Atoi(strings.Split(instruction, "_")[1])
	err = mw.AdaptiveResizeImage(uint(width), height)
	if err != nil {
		return nil, err
	}

	res := mw.GetImageBlob()
	return res, nil
}
