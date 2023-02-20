package client

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type screen interface {
	update()
	draw(screen *ebiten.Image)
}

type packetHandlerScreen interface {
	screen
	handlePacket(packet []byte) error
}
