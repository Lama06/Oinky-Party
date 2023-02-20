package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type gameScreen struct {
	client *client
}

var _ packetHandlerScreen = (*gameScreen)(nil)

func newGameScreen(client *client) *gameScreen {
	return &gameScreen{
		client: client,
	}
}

func (g *gameScreen) update() {
	if g.client.currentGame == nil {
		return
	}

	g.client.currentGame.Update()

	if inpututil.IsKeyJustReleased(ebiten.KeyEscape) {
		endGame, err := json.Marshal(protocol.EndGamePacket{
			PacketName: protocol.EndGamePacketName,
		})
		if err != nil {
			panic(err)
		}
		g.client.SendPacket(endGame)
	}
}

func (g *gameScreen) draw(screen *ebiten.Image) {
	if g.client.currentGame == nil {
		return
	}

	g.client.currentGame.Draw(screen)
}

func (g *gameScreen) handlePacket(packet []byte) error {
	if g.client.currentGame == nil {
		return errors.New("no game")
	}

	err := g.client.currentGame.HandlePacket(packet)
	if err != nil {
		return fmt.Errorf("game failed to handle packet: %w", err)
	}

	return nil
}
