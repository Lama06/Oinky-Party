package client

import (
	"encoding/json"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type createPartyScreen struct {
	client         *client
	partyName      string
	partyNameText  *ui.Text
	continueButton *ui.Button
}

var _ screen = (*createPartyScreen)(nil)

func newCreatePartyScreen(client *client) *createPartyScreen {
	screen := createPartyScreen{
		client:    client,
		partyName: "Neue Party",
		partyNameText: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Name: Neue Party",
			Colors: &ui.TitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
	}

	screen.continueButton = ui.NewButton(ui.ButtonConfig{
		Pos: ui.DynamicPosition(func(width, height int) ui.Position {
			return ui.CenteredPosition{X: width / 2, Y: height / 3 * 2}
		}),
		Text: "Party erstellen",
		Callback: func() {
			createParty, err := json.Marshal(protocol.CreatePartyPacket{
				PacketName: protocol.CreatePartyPacketName,
				Name:       screen.partyName,
			})
			if err != nil {
				panic(err)
			}
			client.SendPacket(createParty)
		},
	})

	return &screen
}

func (c *createPartyScreen) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		c.client.currentScreen = newTitleScreen(c.client)
	}

	c.partyName = string(ebiten.AppendInputChars([]rune(c.partyName)))
	if len(c.partyName) != 0 && inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		c.partyName = c.partyName[:len(c.partyName)-1]
	}
	c.partyNameText.Text = "Name: " + c.partyName

	c.partyNameText.Update()
	c.continueButton.Update()
}

func (c *createPartyScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)
	c.partyNameText.Draw(screen)
	c.continueButton.Draw(screen)
}
