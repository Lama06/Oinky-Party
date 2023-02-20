package ui

import (
	"image/color"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type TextColorPalette struct {
	Color      color.Color
	HoverColor color.Color
}

func (t TextColorPalette) getHoverColor() color.Color {
	if t.HoverColor == nil {
		return t.Color
	}
	return t.HoverColor
}

type TextConfig struct {
	Pos    Position
	Text   string
	Colors *TextColorPalette
	Font   font.Face
}

type Text struct {
	Pos    Position
	Text   string
	Colors *TextColorPalette
	Font   font.Face
}

var _ Component = (*Text)(nil)

func NewText(config TextConfig) *Text {
	return &Text{
		Pos:    config.Pos,
		Text:   config.Text,
		Colors: config.Colors,
		Font:   config.Font,
	}
}

func (t *Text) getColors() TextColorPalette {
	if t.Colors == nil {
		return TextColors
	}
	return *t.Colors
}

func (t *Text) getFont() font.Face {
	if t.Font == nil {
		return rescources.RobotoNormalFont
	}
	return t.Font
}

func (t *Text) Update() {

}

func (t *Text) Draw(screen *ebiten.Image) {
	if t.Pos == nil {
		return
	}

	textBounds := text.BoundString(t.getFont(), t.Text).Size()
	textWidth, textHeight := textBounds.X, textBounds.Y
	topLeftCornerX, topLeftCornerY := t.Pos.TopLeftCorner(textWidth, textHeight)

	color := t.getColors().Color
	if IsHovered(t.Pos, textWidth, textHeight) {
		color = t.getColors().getHoverColor()
	}

	text.Draw(screen, t.Text, t.getFont(), topLeftCornerX, topLeftCornerY+textHeight, color)
}
