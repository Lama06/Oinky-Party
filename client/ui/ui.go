package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Component interface {
	Update()
	Draw(screen *ebiten.Image)
}

type Position interface {
	ToTopLeftPosition(width, height int) TopLeftPosition
	ToCenteredPosition(width, height int) CenteredPosition
}

type TopLeftPosition struct {
	x, y int
}

var _ Position = TopLeftPosition{}

func NewTopLeftPosition(x, y int) TopLeftPosition {
	return TopLeftPosition{
		x: x,
		y: y,
	}
}

func (t TopLeftPosition) X() int {
	return t.x
}

func (t TopLeftPosition) Y() int {
	return t.y
}

func (t TopLeftPosition) ToTopLeftPosition(int, int) TopLeftPosition {
	return t
}

func (t TopLeftPosition) ToCenteredPosition(width, height int) CenteredPosition {
	return CenteredPosition{
		x: t.x + width/2,
		y: t.y + height/2,
	}
}

type CenteredPosition struct {
	x, y int
}

var _ Position = CenteredPosition{}

func NewCenteredPosition(x, y int) CenteredPosition {
	return CenteredPosition{
		x: x,
		y: y,
	}
}

func (c CenteredPosition) X() int {
	return c.x
}

func (c CenteredPosition) Y() int {
	return c.y
}

func (c CenteredPosition) ToTopLeftPosition(width, height int) TopLeftPosition {
	return TopLeftPosition{
		x: c.x - width/2,
		y: c.y - height/2,
	}
}

func (c CenteredPosition) ToCenteredPosition(int, int) CenteredPosition {
	return c
}

func IsInside(pos Position, width, height, x, y int) bool {
	topLeft := pos.ToTopLeftPosition(width, height)

	if x < topLeft.X() || x > topLeft.X()+width {
		return false
	}

	if y < topLeft.Y() || y > topLeft.Y()+height {
		return false
	}

	return true
}

func IsHovered(p Position, width, height int) bool {
	x, y := ebiten.CursorPosition()
	return IsInside(p, width, height, x, y)
}

func IsClicked(p Position, width, height int, btn ebiten.MouseButton) bool {
	return IsHovered(p, width, height) && inpututil.IsMouseButtonJustReleased(btn)
}
