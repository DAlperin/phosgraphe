package transforms

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"regexp"
)

type TransformManager struct {
	Handlers []Transformation
}

func New() *TransformManager {
	tm := TransformManager{
		Handlers: []Transformation{
			heightTransformation{},
			widthTransformation{},
			formatTransformation{},
			rotateTransformation{},
		},
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

func genericHandles(patterns []string, instruction instructions.Instruction) (bool, error) {
	for _, s := range patterns {
		match, err := regexp.MatchString(s, instruction)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
