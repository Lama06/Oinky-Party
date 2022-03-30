package ui

import (
	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
)

const (
	buttonPadding = 20
)

var (
	buttonBackgroundColor      = colornames.Aqua
	buttonTextColor            = colornames.Black
	buttonHoverBackgroundColor = colornames.Blue
	buttonHoverTextColor       = colornames.Snow
)

type Button struct {
	pos      Position
	text     string
	callback func()
}

var _ Component = (*Button)(nil)

func NewButton(pos Position, text string, callback func()) *Button {
	return &Button{
		pos:      pos,
		text:     text,
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

	bgColor := buttonBackgroundColor
	textColor := buttonTextColor
	if IsHovered(topLeftPosition, buttonWidth, buttonHeight) {
		bgColor = buttonHoverBackgroundColor
		textColor = buttonHoverTextColor
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
