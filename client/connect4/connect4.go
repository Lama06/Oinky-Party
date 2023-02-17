package connect4

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"

	"github.com/Lama06/Oinky-Party/client/game"
	shared "github.com/Lama06/Oinky-Party/connect4"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

const cellSize = 50

type board [shared.BoardWidth][shared.BoardHeight]shared.Cell

func (b *board) place(color shared.Color, x int) {
	y := shared.BoardHeight - 1
	for b[x][y] != shared.EmptyCell {
		y--
		if y == -1 {
			return
		}
	}
	b[x][y] = color.ToCell()
}

type impl struct {
	client game.Client
	board  *board
}

var _ game.Game = (*impl)(nil)

func create(client game.Client) game.Game {
	return &impl{
		client: client,
		board:  &board{},
	}
}

var _ game.Creator = create

func (i *impl) HandleGameStarted() {}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePacket(data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	switch packetName {
	case shared.PlayerPlacedPacketName:
		var playerPlaced shared.PlayerPlacedPacket
		err := json.Unmarshal(data, &playerPlaced)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		i.board.place(playerPlaced.Player, int(playerPlaced.X))
		return nil
	default:
		return errors.New("unknown packet name")
	}
}

func (i *impl) Draw(screen *ebiten.Image) {
	for x := 0; x < shared.BoardWidth; x++ {
		for y := 0; y < shared.BoardHeight; y++ {
			var clr color.Color
			switch i.board[x][y] {
			case shared.EmptyCell:
				continue
			case shared.RedCell:
				clr = colornames.Red
			case shared.YellowCell:
				clr = colornames.Yellow
			}
			ebitenutil.DrawRect(screen, float64(x*cellSize), float64(y*cellSize), cellSize, cellSize, clr)
		}
	}
}

func (i *impl) Update() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mouseX, _ := ebiten.CursorPosition()
		x := mouseX / cellSize
		if x < 0 || x > shared.BoardWidth-1 {
			return
		}

		place, err := json.Marshal(shared.PlacePacket{
			PacketName: shared.PlacePacketName,
			X:          int32(x),
		})
		if err != nil {
			panic(err)
		}
		i.client.SendPacket(place)
	}
}

func (i *impl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return shared.BoardWidth * cellSize, shared.BoardHeight * cellSize
}

var Type = game.Type{
	Creator:     create,
	Name:        shared.Name,
	DisplayName: "Vier Gewinnt",
}
