package connect4

import (
	"encoding/json"
	"errors"
	"fmt"

	shared "github.com/Lama06/Oinky-Party/connect4"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
)

type board [shared.BoardWidth][shared.BoardHeight]shared.Cell

func getWinner(cells []shared.Cell) (winner shared.Color, found bool) {
	var currentColor shared.Color
	var hasCurrentColor bool
	var currentCount int
	for _, cell := range cells {
		if cell == shared.EmptyCell {
			hasCurrentColor = false
			currentCount = 0
			continue
		}
		if !hasCurrentColor || cell.ToColor() != currentColor {
			currentColor = cell.ToColor()
			hasCurrentColor = true
			currentCount = 1
			continue
		}
		currentCount++
		if currentCount == 4 {
			return currentColor, true
		}
	}

	return shared.RedColor, false
}

func (b *board) getWinnerInRow(y int) (winner shared.Color, found bool) {
	cells := make([]shared.Cell, shared.BoardWidth)
	for x := 0; x < shared.BoardWidth; x++ {
		cells[x] = b[x][y]
	}
	return getWinner(cells)
}

func (b *board) getWinnerInColumn(x int) (winner shared.Color, found bool) {
	cells := make([]shared.Cell, shared.BoardHeight)
	for y := 0; y < shared.BoardHeight; y++ {
		cells[y] = b[x][y]
	}
	return getWinner(cells)
}

func (b *board) getWinner() (winner shared.Color, found bool) {
	for y := 0; y < shared.BoardHeight; y++ {
		if winner, found := b.getWinnerInRow(y); found {
			return winner, true
		}
	}

	for x := 0; x < shared.BoardHeight; x++ {
		if winner, found := b.getWinnerInColumn(x); found {
			return winner, true
		}
	}

	return shared.RedColor, false
}

func (b *board) canPlace(x int) bool {
	return b[x][0] == shared.EmptyCell
}

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
	party         game.Party
	board         *board
	red           game.Player
	yellow        game.Player
	currentPlayer shared.Color
}

var _ game.Game = (*impl)(nil)

func create(party game.Party) game.Game {
	var players []game.Player
	for _, player := range party.Players() {
		players = append(players, player)
	}
	if len(players) != 2 {
		return nil
	}

	return &impl{
		party:         party,
		board:         &board{},
		red:           players[0],
		yellow:        players[1],
		currentPlayer: shared.RedColor,
	}
}

var _ game.Creator = create

func (i *impl) getColor(player game.Player) shared.Color {
	switch player {
	case i.red:
		return shared.RedColor
	case i.yellow:
		return shared.YellowColor
	default:
		panic("invalid player")
	}
}

func (i *impl) HandleGameStarted() {}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePlayerLeft(player game.Player) {
	i.party.EndGame()
}

func (i *impl) HandlePacket(sender game.Player, data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to get packet name: %w", err)
	}
	switch packetName {
	case shared.PlacePacketName:
		var playerPlaced shared.PlayerPlacedPacket
		err := json.Unmarshal(data, &playerPlaced)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		if playerPlaced.X < 0 || playerPlaced.X > shared.BoardWidth-1 {
			return fmt.Errorf("invalid column: %d", playerPlaced.X)
		}
		if i.getColor(sender) != i.currentPlayer {
			return errors.New("its not this players turn")
		}
		if !i.board.canPlace(int(playerPlaced.X)) {
			return fmt.Errorf("cannot place in column: %d", playerPlaced.X)
		}

		i.board.place(i.getColor(sender), int(playerPlaced.X))
		i.currentPlayer = !i.currentPlayer

		place, err := json.Marshal(shared.PlayerPlacedPacket{
			PacketName: shared.PlayerPlacedPacketName,
			Player:     i.getColor(sender),
			X:          playerPlaced.X,
		})
		if err != nil {
			panic(err)
		}
		i.party.BroadcastPacket(place)

		if _, found := i.board.getWinner(); found {
			i.party.EndGame()
		}

		return nil
	default:
		return errors.New("unknown packet name")
	}
}

func (i *impl) Tick() {}

var Type = game.Type{
	Creator: create,
	Name:    shared.Name,
}
