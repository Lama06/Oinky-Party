package client

import (
	"encoding/json"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type changeNameScreen struct {
	client         *client
	newName        string
	newNameText    *ui.Text
	continueButton *ui.Button
}

var _ screen = (*changeNameScreen)(nil)

func newChangeNameScreen(client *client) *changeNameScreen {
	screen := changeNameScreen{
		client:  client,
		newName: client.name,
		newNameText: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   client.name,
			Colors: &ui.TitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
	}

	screen.continueButton = ui.NewButton(ui.ButtonConfig{
		Pos: ui.DynamicPosition(func(width, height int) ui.Position {
			return ui.CenteredPosition{X: width / 2, Y: (height / 3) * 2}
		}),
		Text: "Namen ändern",
		Callback: func() {
			changeName, err := json.Marshal(protocol.ChangeNamePacket{
				PacketName: protocol.ChangeNamePacketName,
				NewName:    screen.newName,
			})
			if err != nil {
				panic(err)
			}
			client.SendPacket(changeName)

			client.name = screen.newName

			client.currentScreen = newTitleScreen(client)
		},
	})

	return &screen
}

func (c *changeNameScreen) components() []ui.Component {
	return []ui.Component{c.newNameText, c.continueButton}
}

func (c *changeNameScreen) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		c.client.currentScreen = newTitleScreen(c.client)
	}

	c.newName = string(ebiten.AppendInputChars([]rune(c.newName)))
	if len(c.newName) != 0 && inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		c.newName = c.newName[:len(c.newName)-1]
	}
	c.newNameText.Text = c.newName

	for _, component := range c.components() {
		component.Update()
	}
}

func (c *changeNameScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)

	for _, component := range c.components() {
		component.Draw(screen)
	}
}
