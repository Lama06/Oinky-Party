package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Position interface {
	TopLeftCorner(width, height int) (int, int)
}

func IsInside(pos Position, width, height, x, y int) bool {
	topLeftX, topLeftY := pos.TopLeftCorner(width, height)

	if x < topLeftX || x > topLeftX+width {
		return false
	}

	if y < topLeftY || y > topLeftY+height {
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

type TopLeftCornerPosition struct {
	X, Y int
}

var _ Position = TopLeftCornerPosition{}

func (t TopLeftCornerPosition) TopLeftCorner(int, int) (int, int) {
	return t.X, t.Y
}

type CenteredPosition struct {
	X, Y int
}

var _ Position = CenteredPosition{}

func (c CenteredPosition) TopLeftCorner(width, height int) (int, int) {
	return c.X - width/2, c.Y - height/2
}

type DynamicPosition func(width, height int) Position

var _ Position = DynamicPosition(nil)

func (d DynamicPosition) TopLeftCorner(width, height int) (int, int) {
	return d(ebiten.WindowSize()).TopLeftCorner(width, height)
}
