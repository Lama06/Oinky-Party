package client

import (
	"encoding/json"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type partyScreenPlayerName struct {
	id   int32
	text *ui.Text
}

type partyScreen struct {
	client          *client
	title           *ui.Text
	playersNames    []partyScreenPlayerName
	startGameButton *ui.Button
}

var _ screen = (*partyScreen)(nil)

func newPartyScreen(client *client) *partyScreen {
	return &partyScreen{
		client: client,
		title: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Party: " + client.partyName,
			Colors: &ui.TitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
		startGameButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height - 100}
			}),
			Text: "Spiel starten",
			Callback: func() {
				client.currentScreen = newStartGameScreen(client)
			},
		}),
	}
}

func (p *partyScreen) arePlayerNamesValid() bool {
	players := p.client.partyPlayersSorted()
	if len(players) != len(p.playersNames) {
		return false
	}
	for i, player := range players {
		if p.playersNames[i].id != player.Id {
			return false
		}
	}
	return true
}

func (p *partyScreen) updatePlayerList() {
	players := p.client.partyPlayersSorted()

	p.playersNames = make([]partyScreenPlayerName, len(players))
	for i, player := range players {
		iCopy := i
		p.playersNames[i] = partyScreenPlayerName{
			id: player.Id,
			text: ui.NewText(ui.TextConfig{
				Pos: ui.DynamicPosition(func(width, height int) ui.Position {
					return ui.CenteredPosition{X: width / 2, Y: 100 + height/3 + 100*iCopy}
				}),
				Text: player.Name,
			}),
		}
	}
}

func (p *partyScreen) components() []ui.Component {
	components := make([]ui.Component, 0)
	components = append(components, p.title, p.startGameButton)

	for _, playerName := range p.playersNames {
		components = append(components, playerName.text)
	}

	return components
}

func (p *partyScreen) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		leaveParty, err := json.Marshal(protocol.LeavePartyPacket{
			PacketName: protocol.LeavePartyPacketName,
		})
		if err != nil {
			panic(err)
		}
		p.client.SendPacket(leaveParty)
	}

	if !p.arePlayerNamesValid() {
		p.updatePlayerList()
	}

	for _, component := range p.components() {
		component.Update()
	}
}

func (p *partyScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)
	for _, component := range p.components() {
		component.Draw(screen)
	}
}
