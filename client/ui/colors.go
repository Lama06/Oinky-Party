package ui

import "image/color"

var (
	DefaultBackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	DefaultButtonColors    = ButtonColorPalette{
		BackgroundColor:      color.RGBA{R: 18, G: 53, B: 91, A: 255},
		BackgroundHoverColor: color.RGBA{R: 134, G: 22, B: 87, A: 255},
		TextColor:            color.RGBA{R: 212, G: 245, B: 245, A: 255},
		TextHoverColor:       color.RGBA{R: 212, G: 245, B: 245, A: 255},
	}
	DefaultTextColors = TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 82, G: 73, B: 72, A: 255},
	}
	DefaultTitleColors = TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 112, G: 248, B: 186, A: 255},
	}
)
