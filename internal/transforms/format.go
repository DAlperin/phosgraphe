package transforms

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"gopkg.in/gographics/imagick.v3/imagick"
	"net/http"
	"strings"
)

type formatTransformation struct {
}

func (fT formatTransformation) Handles(transform string) (bool, error) {
	accepted := []string{"format_[A-Z,a-z]+"}
	return genericHandles(accepted, transform)
}

func (fT formatTransformation) Transform(file []byte, instruction instructions.Instruction) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImageBlob(file)
	if err != nil {
		return nil, err
	}

	format := strings.Split(instruction, "_")[1]

	currentFormat := strings.Split(http.DetectContentType(file), "/")[1]

	if currentFormat == format {
		fmt.Println("Already in requested format, ignoring")
		return file, nil
	}

	err = mw.SetImageFormat(format)
	if err != nil {
		return nil, err
	}

	res := mw.GetImageBlob()
	return res, nil
}
