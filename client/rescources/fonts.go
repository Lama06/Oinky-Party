package rescources

import (
	_ "embed"
	"fmt"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	//go:embed roboto_bold.ttf
	robotoFontData   []byte
	RobotoNormalFont font.Face
	RobotoTitleFont  font.Face
)

func init() {
	const errorMsgFormat = "failed to instantiate roboto font face: %w"

	robotoFont, err := opentype.Parse(robotoFontData)
	if err != nil {
		panic(fmt.Errorf("failed to partse roboto font: %w", err))
	}

	RobotoNormalFont, err = opentype.NewFace(robotoFont, &opentype.FaceOptions{
		Size:    28,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(fmt.Errorf(errorMsgFormat, err))
	}

	RobotoTitleFont, err = opentype.NewFace(robotoFont, &opentype.FaceOptions{
		Size:    70,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(fmt.Errorf(errorMsgFormat, err))
	}
}
