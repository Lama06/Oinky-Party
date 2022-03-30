package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

var (
	textColor      = colornames.White
	textHoverColor = colornames.Snow
)

type Text struct {
	pos  Position
	text string
	font font.Face
}

var _ Component = (*Text)(nil)

func NewText(pos Position, text string, font font.Face) *Text {
	return &Text{
		pos:  pos,
		text: text,
		font: font,
	}
}

func (t *Text) Update() {}

func (t *Text) Draw(screen *ebiten.Image) {
	textBounds := text.BoundString(t.font, t.text).Size()
	textWidth, textHeight := textBounds.X, textBounds.Y
	topLeftPosition := t.pos.ToTopLeftPosition(textWidth, textHeight)

	color := textColor
	if IsHovered(topLeftPosition, textWidth, textHeight) {
		color = textHoverColor
	}

	text.Draw(screen, t.text, t.font, topLeftPosition.x, topLeftPosition.y+textHeight, color)
}
