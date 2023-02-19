package schiffe_versenken

import (
	"encoding/json"

	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type enemyBoard struct {
	game    *impl
	ships   [shared.BoardWidth][shared.BoardHeight]bool
	markers [shared.BoardWidth][shared.BoardHeight]bool
}

func newEmptyEnemyBoard(game *impl) *enemyBoard {
	return &enemyBoard{
		game: game,
	}
}

func (e *enemyBoard) draw(screen *ebiten.Image) {
	board := ebiten.NewImage(boardWidth, boardHeight)
	drawBorders(board)
	drawShips(board, e.ships)
	drawMarkers(board, e.markers)
	var boardDrawOptions ebiten.DrawImageOptions
	boardDrawOptions.GeoM.Translate(boardWidth+distanceBetweenBoards, 0)
	screen.DrawImage(board, &boardDrawOptions)
}

func (e *enemyBoard) update() {
	mouseX, mouseY := ebiten.CursorPosition()

	fieldX, fieldY, ok := getFieldCoordinates(mouseX-boardWidth-distanceBetweenBoards, mouseY)
	if !ok {
		return
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		e.markers[fieldX][fieldY] = !e.markers[fieldX][fieldY]
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		fire, err := json.Marshal(shared.FirePacket{
			PacketName: shared.FirePacketName,
			Position:   shared.Position{X: fieldX, Y: fieldY},
		})
		if err != nil {
			panic(err)
		}
		e.game.client.SendPacket(fire)
	}
}

func (e *enemyBoard) handleFireResultPacket(packet shared.FireResultPacket) {
	e.markers[packet.Position.X][packet.Position.Y] = true

	if packet.Hit {
		e.ships[packet.Position.X][packet.Position.Y] = true
	}
}
