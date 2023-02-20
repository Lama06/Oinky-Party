package schiffe_versenken

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Lama06/Oinky-Party/client/game"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

const (
	fieldSize                 = 50
	borderWidth               = 1
	distanceBetweenBoards     = 80
	numberOfHorizontalBorders = shared.BoardHeight + 1
	numberOfVerticalBorders   = shared.BoardWidth + 1
	boardWidth                = shared.BoardWidth*fieldSize + numberOfVerticalBorders*borderWidth
	boardHeight               = shared.BoardHeight*fieldSize + numberOfHorizontalBorders*borderWidth
)

type impl struct {
	client                    game.Client
	setupShipsContinueBtn     *ui.Button
	setupBoard                *setupBoard
	hasSetupShips             bool
	waitingForGameToStartText *ui.Text
	gameStarted               bool
	personalBoard             *personalBoard
	enemyBoard                *enemyBoard
}

var _ game.Game = (*impl)(nil)

func create(client game.Client) game.Game {
	return &impl{
		client: client,
	}
}

var _ game.Creator = create

func (i *impl) HandleGameStarted() {
	i.setupShipsContinueBtn = i.createSetupSetupShipsContinueBtn()
	i.setupBoard = newEmptySetupBoard(i)
	i.waitingForGameToStartText = i.createWaitingForGameToStartText()
	i.enemyBoard = newEmptyEnemyBoard(i)
}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePacket(data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	switch packetName {
	case shared.GameStartedPacketName:
		if !i.hasSetupShips {
			return errors.New("game started before player set up their ships")
		}

		i.gameStarted = true
		return nil
	case shared.FireResultPacketName:
		if !i.gameStarted {
			return errors.New("game has not started yet")
		}

		var fireResult shared.FireResultPacket
		err := json.Unmarshal(data, &fireResult)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		if !fireResult.Position.Valid() {
			return errors.New("invalid position")
		}

		i.enemyBoard.handleFireResultPacket(fireResult)

		return nil
	case shared.OpponentFiredPacketName:
		if !i.gameStarted {
			return errors.New("game has not started yet")
		}

		var opponentFired shared.OpponentFiredPacket
		err := json.Unmarshal(data, &opponentFired)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}

		if !opponentFired.Position.Valid() {
			return errors.New("invalid position")
		}

		i.personalBoard.handleOponentFiredPacket(opponentFired)

		return nil
	default:
		return fmt.Errorf("unknown packet name: %s", packetName)
	}
}

func (i *impl) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.White)

	if !i.hasSetupShips {
		i.setupBoard.draw(screen)
		i.setupShipsContinueBtn.Draw(screen)
	} else if !i.gameStarted {
		i.waitingForGameToStartText.Draw(screen)
	} else {
		i.personalBoard.draw(screen)
		i.enemyBoard.draw(screen)
	}
}

func (i *impl) Update() {
	if !i.hasSetupShips {
		i.setupBoard.update()
		switch i.setupBoard.parseShips().Valid() {
		case false:
			i.setupShipsContinueBtn.SetColors(&ui.DisabledButtonColors)
		case true:
			i.setupShipsContinueBtn.SetColors(&ui.ButtonColors)
		}
		i.setupShipsContinueBtn.Update()
	} else if !i.gameStarted {
		i.waitingForGameToStartText.Update()
	} else {
		i.enemyBoard.update()
	}
}

func (i *impl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if !i.hasSetupShips {
		return boardWidth + 200, boardHeight
	} else if !i.gameStarted {
		return outsideWidth, outsideHeight
	} else {
		return boardWidth*2 + distanceBetweenBoards, boardHeight
	}
}

func (i *impl) createSetupSetupShipsContinueBtn() *ui.Button {
	return ui.NewButton(ui.ButtonConfig{
		Pos:  ui.CenteredPosition{X: boardWidth + 100, Y: boardHeight / 2},
		Text: "Weiter",
		Callback: func() {
			if i.hasSetupShips {
				return
			}

			ships := i.setupBoard.parseShips()

			if !ships.Valid() {
				return
			}

			i.hasSetupShips = true
			i.personalBoard = newPersonalBoard(i, ships)

			setupShips, err := json.Marshal(shared.SetupShipsPacket{
				PacketName: shared.SetupShipsPacketName,
				Ships:      ships,
			})
			if err != nil {
				panic(err)
			}
			i.client.SendPacket(setupShips)
		},
	})
}

func (i *impl) createWaitingForGameToStartText() *ui.Text {
	return ui.NewText(ui.TextConfig{
		Pos: ui.DynamicPosition(func(width, height int) ui.Position {
			return ui.CenteredPosition{X: width / 2, Y: height / 2}
		}),
		Text: "Warten auf Gegner...",
	})
}

var Type = game.Type{
	Name:        shared.Name,
	DisplayName: "Schiffe versenken",
	Creator:     create,
}
