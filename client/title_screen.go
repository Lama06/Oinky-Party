package client

import (
	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type titleScreen struct {
	client            *client
	title             *ui.Text
	createPartyButton *ui.Button
	joinPartyButton   *ui.Button
	changeNameButton  *ui.Button
}

var _ screen = (*titleScreen)(nil)

func newTitleScreen(client *client) *titleScreen {
	return &titleScreen{
		client: client,
		title: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Oinky Party",
			Colors: &ui.TitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
		createPartyButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: (height / 3) * 2}
			}),
			Text: "Party erstellen",
			Callback: func() {
				client.currentScreen = newCreatePartyScreen(client)
			},
		}),
		joinPartyButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: (height/3)*2 + 100}
			}),
			Text: "Party beitreten",
			Callback: func() {
				client.currentScreen = newJoinPartyScreenLoading(client)
			},
		}),
		changeNameButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: (height/3)*2 + 200}
			}),
			Text: "Namen ändern",
			Callback: func() {
				client.currentScreen = newChangeNameScreen(client)
			},
		}),
	}
}

func (t *titleScreen) components() []ui.Component {
	return []ui.Component{t.title, t.createPartyButton, t.joinPartyButton, t.changeNameButton}
}

func (t *titleScreen) update() {
	if inpututil.IsKeyJustReleased(ebiten.Key1) {
		t.client.currentScreen = newCreatePartyScreen(t.client)
	} else if inpututil.IsKeyJustReleased(ebiten.Key2) {
		t.client.currentScreen = newJoinPartyScreenLoading(t.client)
	} else if inpututil.IsKeyJustReleased(ebiten.Key3) {
		t.client.currentScreen = newChangeNameScreen(t.client)
	}

	for _, component := range t.components() {
		component.Update()
	}
}

func (t *titleScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)
	for _, component := range t.components() {
		component.Draw(screen)
	}
}
