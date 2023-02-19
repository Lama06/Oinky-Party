package schiffe_versenken

import (
	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/hajimehoshi/ebiten/v2"
)

type personalBoard struct {
	game  *impl
	ships [shared.BoardWidth][shared.BoardHeight]bool
	hits  [shared.BoardWidth][shared.BoardHeight]bool
}

func newPersonalBoard(game *impl, ships shared.Ships) *personalBoard {
	board := personalBoard{
		game: game,
	}

	for _, ship := range ships {
		for _, field := range ship {
			board.ships[field.X][field.Y] = true
		}
	}

	return &board
}

func (p *personalBoard) handleOponentFiredPacket(packet shared.OpponentFiredPacket) {
	p.hits[packet.Position.X][packet.Position.Y] = true
}

func (p *personalBoard) draw(screen *ebiten.Image) {
	board := ebiten.NewImage(boardWidth, boardHeight)
	drawBorders(board)
	drawShips(board, p.ships)
	drawMarkers(board, p.hits)
	screen.DrawImage(board, nil)
}
