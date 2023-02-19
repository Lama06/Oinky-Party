package schiffe_versenken

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Lama06/Oinky-Party/protocol"
	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/Lama06/Oinky-Party/server/game"
)

type cell byte

const (
	emptyCell cell = iota
	shipCell
)

type board [shared.BoardWidth][shared.BoardHeight]cell

func newBoardFromShips(ships shared.Ships) *board {
	result := &board{}
	for _, ship := range ships {
		for _, pos := range ship {
			result[pos.X][pos.Y] = shipCell
		}
	}
	return result
}

func (b *board) isEmpty() bool {
	for x := 0; x < shared.BoardWidth; x++ {
		for y := 0; y < shared.BoardHeight; y++ {
			if b[x][y] == shipCell {
				return false
			}
		}
	}

	return true
}

func (b *board) fire(pos shared.Position) (hit bool) {
	hit = b[pos.X][pos.Y] == shipCell
	b[pos.X][pos.Y] = emptyCell
	return
}

type player struct {
	handle        game.Player
	hasSetupShips bool
	board         *board
}

func newPlayer(handle game.Player) *player {
	return &player{
		handle: handle,
	}
}

type impl struct {
	party         game.Party
	gameStarted   bool
	player1       *player
	player2       *player
	currentPlayer *player
}

var _ game.Game = (*impl)(nil)

func create(party game.Party) game.Game {
	if len(party.Players()) != 2 {
		return nil
	}

	players := make([]game.Player, 0, 2)
	for _, player := range party.Players() {
		players = append(players, player)
	}

	player1 := newPlayer(players[0])
	player2 := newPlayer(players[1])

	return &impl{
		party:         party,
		player1:       player1,
		player2:       player2,
		currentPlayer: player1,
	}
}

var _ game.Creator = create

func (i *impl) HandleGameStarted() {}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePlayerLeft(player game.Player) {
	i.party.EndGame()
}

func (i *impl) HandlePacket(sender game.Player, data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to obtain the packet name: %w", err)
	}

	senderPlayer := i.getPlayer(sender)

	var otherPlayer *player
	switch senderPlayer {
	case i.player1:
		otherPlayer = i.player2
	case i.player2:
		otherPlayer = i.player1
	}

	switch packetName {
	case shared.SetupShipsPacketName:
		var setupShips shared.SetupShipsPacket
		err := json.Unmarshal(data, &setupShips)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		if !setupShips.Ships.Valid() {
			return errors.New("invalid ships")
		}

		if senderPlayer.hasSetupShips {
			return errors.New("player has already setup their ships")
		}

		senderPlayer.hasSetupShips = true
		senderPlayer.board = newBoardFromShips(setupShips.Ships)

		if otherPlayer.hasSetupShips {
			i.gameStarted = true

			gameStarted, err := json.Marshal(shared.GameStartedPacket{
				PacketName: shared.GameStartedPacketName,
			})
			if err != nil {
				panic(err)
			}
			i.party.BroadcastPacket(gameStarted)
		}

		return nil
	case shared.FirePacketName:
		var fire shared.FirePacket
		err := json.Unmarshal(data, &fire)
		if err != nil {
			return fmt.Errorf("failed to unmarshl json: %w", err)
		}

		if !fire.Position.Valid() {
			return errors.New("invalid position")
		}

		if !i.gameStarted {
			return errors.New("game has not started yet")
		}

		if senderPlayer != i.currentPlayer {
			return errors.New("its not your turn")
		}

		hit := otherPlayer.board.fire(fire.Position)

		if hit && otherPlayer.board.isEmpty() {
			i.party.EndGame()
			return nil
		}

		fireResult, err := json.Marshal(shared.FireResultPacket{
			PacketName: shared.FireResultPacketName,
			Position:   fire.Position,
			Hit:        hit,
		})
		if err != nil {
			panic(err)
		}
		sender.SendPacket(fireResult)

		opponentFired, err := json.Marshal(shared.OpponentFiredPacket{
			PacketName: shared.OpponentFiredPacketName,
			Position:   fire.Position,
		})
		if err != nil {
			panic(err)
		}
		otherPlayer.handle.SendPacket(opponentFired)

		if !hit {
			i.currentPlayer = otherPlayer
		}

		return nil
	default:
		return fmt.Errorf("unknown packet name: %s", packetName)
	}
}

func (i *impl) Tick() {}

func (i *impl) getPlayer(player game.Player) *player {
	switch player {
	case i.player1.handle:
		return i.player1
	case i.player2.handle:
		return i.player2
	default:
		return nil
	}
}

var Type = game.Type{
	Name:    shared.Name,
	Creator: create,
}
