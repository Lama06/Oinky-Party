package client

import (
	"encoding/json"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type startGameScreen struct {
	client      *client
	title       *ui.Text
	gameButtons []*ui.Button
}

var _ screen = (*startGameScreen)(nil)

func newStartGameScreen(client *client) *startGameScreen {
	gameButtons := make([]*ui.Button, len(gameTypes))
	for i, gameType := range gameTypes {
		iCopy := i
		gameTypeCopy := gameType

		gameButtons[i] = ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: 100 + height/3 + 100*iCopy}
			}),
			Text: gameType.DisplayName,
			Callback: func() {
				startGame, err := json.Marshal(protocol.StartGamePacket{
					PacketName: protocol.StartGamePacketName,
					GameType:   gameTypeCopy.Name,
				})
				if err != nil {
					panic(err)
				}
				client.SendPacket(startGame)
			},
		})
	}

	return &startGameScreen{
		client: client,
		title: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Spiel starten",
			Colors: &ui.TitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
		gameButtons: gameButtons,
	}
}

func (s *startGameScreen) components() []ui.Component {
	components := []ui.Component{s.title}

	for _, button := range s.gameButtons {
		components = append(components, button)
	}

	return components
}

func (s *startGameScreen) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.client.currentScreen = newPartyScreen(s.client)
	}

	for _, component := range s.components() {
		component.Update()
	}
}

func (s *startGameScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)
	for _, component := range s.components() {
		component.Draw(screen)
	}
}
