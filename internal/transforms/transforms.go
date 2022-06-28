package transforms

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"gopkg.in/gographics/imagick.v3/imagick"
)

type TransformManager struct {
	Handlers    []Transformation
	ImageMagick *imagick.MagickWand
}

func New() *TransformManager {
	mw := imagick.NewMagickWand()

	tm := TransformManager{
		Handlers: []Transformation{
			heightTransformation{
				ImageMagick: mw,
			},
			widthTransformation{
				ImageMagick: mw,
			},
		},
		ImageMagick: mw,
	}

	return &tm
}

func (tm *TransformManager) GetHandler(instruction string) (Transformation, error) {
	for _, k := range tm.Handlers {
		handles, err := k.Handles(instruction)
		if err != nil {
			return nil, err
		}
		if handles {
			return k, nil
		}
	}
	return nil, fmt.Errorf("cant find handler to handle %s", instruction)
}

type Transformation interface {
	Transform(file []byte, instruction instructions.Instruction) ([]byte, error)
	Handles(transform string) (bool, error)
}
