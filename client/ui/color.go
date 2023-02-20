package ui

import "image/color"

var (
	BackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ButtonColors    = ButtonColorPalette{
		BackgroundColor:      color.RGBA{R: 18, G: 53, B: 91, A: 255},
		BackgroundHoverColor: color.RGBA{R: 134, G: 22, B: 87, A: 255},
		TextColor:            color.RGBA{R: 212, G: 245, B: 245, A: 255},
		TextHoverColor:       color.RGBA{R: 212, G: 245, B: 245, A: 255},
	}
	DisabledButtonColors = ButtonColorPalette{
		BackgroundColor:      color.RGBA{R: 42, G: 59, B: 82, A: 255},
		BackgroundHoverColor: color.RGBA{R: 29, G: 37, B: 48, A: 255},
		TextColor:            color.RGBA{R: 212, G: 245, B: 245, A: 255},
	}
	TextColors = TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 82, G: 73, B: 72, A: 255},
	}
	TitleColors = TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 112, G: 248, B: 186, A: 255},
	}
)
