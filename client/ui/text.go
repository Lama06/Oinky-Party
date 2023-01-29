package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type TextColorPalette struct {
	Color      color.Color
	HoverColor color.Color
}

type Text struct {
	pos   Position
	text  string
	color TextColorPalette
	font  font.Face
}

var _ Component = (*Text)(nil)

func NewText(pos Position, text string, color TextColorPalette, font font.Face) *Text {
	return &Text{
		pos:   pos,
		text:  text,
		color: color,
		font:  font,
	}
}

func (t *Text) Update() {}

func (t *Text) Draw(screen *ebiten.Image) {
	textBounds := text.BoundString(t.font, t.text).Size()
	textWidth, textHeight := textBounds.X, textBounds.Y
	topLeftPosition := t.pos.ToTopLeftPosition(textWidth, textHeight)

	color := t.color.Color
	if IsHovered(topLeftPosition, textWidth, textHeight) && t.color.HoverColor != nil {
		color = t.color.HoverColor
	}

	text.Draw(screen, t.text, t.font, topLeftPosition.x, topLeftPosition.y+textHeight, color)
}
