package ui

import (
	"image/color"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	buttonPadding = 20
)

type ButtonColorPalette struct {
	BackgroundColor      color.Color
	BackgroundHoverColor color.Color
	TextColor            color.Color
	TextHoverColor       color.Color
}

type Button struct {
	pos      Position
	text     string
	color    ButtonColorPalette
	callback func()
}

var _ Component = (*Button)(nil)

func NewButton(pos Position, text string, color ButtonColorPalette, callback func()) *Button {
	return &Button{
		pos:      pos,
		text:     text,
		color:    color,
		callback: callback,
	}
}

func (b *Button) getDimensions() (width int, height int) {
	bounds := text.BoundString(rescources.RobotoNormalFont, b.text)
	textWidth, textHeight := bounds.Size().X, bounds.Size().Y
	return textWidth + buttonPadding*2, textHeight + buttonPadding*2
}

func (b *Button) Update() {
	buttonWidth, buttonHeight := b.getDimensions()
	if IsClicked(b.pos, buttonWidth, buttonHeight, ebiten.MouseButtonLeft) {
		b.callback()
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	buttonWidth, buttonHeight := b.getDimensions()
	topLeftPosition := b.pos.ToTopLeftPosition(buttonWidth, buttonHeight)

	bgColor := b.color.BackgroundColor
	textColor := b.color.TextColor
	if IsHovered(topLeftPosition, buttonWidth, buttonHeight) {
		if b.color.BackgroundHoverColor != nil {
			bgColor = b.color.BackgroundHoverColor
		}
		if b.color.TextHoverColor != nil {
			textColor = b.color.TextHoverColor
		}
	}

	img := ebiten.NewImage(buttonWidth, buttonHeight)
	img.Fill(bgColor)
	textX := buttonPadding
	textY := buttonHeight - buttonPadding
	text.Draw(img, b.text, rescources.RobotoNormalFont, textX, textY, textColor)
	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Translate(float64(topLeftPosition.x), float64(topLeftPosition.y))
	screen.DrawImage(img, &drawOptions)
}
