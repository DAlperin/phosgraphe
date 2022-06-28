package transforms

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"gopkg.in/gographics/imagick.v3/imagick"
	"regexp"
	"strconv"
	"strings"
)

type heightTransformation struct {
	ImageMagick *imagick.MagickWand
}

type widthTransformation struct {
	ImageMagick *imagick.MagickWand
}

func (eT heightTransformation) Handles(transform string) (bool, error) {
	fmt.Println("checking handles")
	accepted := []string{"height_\\d+"}
	for _, s := range accepted {
		match, err := regexp.MatchString(s, transform)
		fmt.Println("a match?", match)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (eT heightTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	fmt.Println("transforming")
	defer eT.ImageMagick.Clear()

	err := eT.ImageMagick.ReadImageBlob(file)
	if err != nil {
		fmt.Println("failed to read image blob?")
		return nil, err
	}

	width := eT.ImageMagick.GetImageWidth()
	fmt.Println(width)
	height, err := strconv.Atoi(strings.Split(string(instruction), "_")[1])

	err = eT.ImageMagick.AdaptiveResizeImage(width, uint(height))
	if err != nil {
		return nil, err
	}

	res := eT.ImageMagick.GetImageBlob()
	return res, nil
}

func (wT widthTransformation) Handles(transform string) (bool, error) {
	fmt.Println("checking handles")
	accepted := []string{"width_\\d+"}
	for _, s := range accepted {
		match, err := regexp.MatchString(s, transform)
		fmt.Println("a match?", match)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (wT widthTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	fmt.Println("transforming")
	defer wT.ImageMagick.Clear()

	err := wT.ImageMagick.ReadImageBlob(file)
	if err != nil {
		fmt.Println("failed to read image blob?")
		return nil, err
	}

	height := wT.ImageMagick.GetImageHeight()

	width, err := strconv.Atoi(strings.Split(string(instruction), "_")[1])

	err = wT.ImageMagick.AdaptiveResizeImage(uint(width), height)
	if err != nil {
		return nil, err
	}

	res := wT.ImageMagick.GetImageBlob()
	return res, nil
}
