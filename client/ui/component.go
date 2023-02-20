package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Component interface {
	Update()
	Draw(screen *ebiten.Image)
}
