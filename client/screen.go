package client

import (
	"encoding/json"
	"fmt"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type screen interface {
	update()
	draw(screen *ebiten.Image)
}

type packetHandlerScreen interface {
	screen
	handlePacket(packet []byte) error
}

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
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text: "Oinky Party",
			Font: rescources.RobotoTitleFont,
		}),
		createPartyButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: (height / 3) * 2}
			}),
			Text: "Party erstellen",
			Callback: func() {
				client.currentScreen = newCreatePartyScreen(client)
			},
		}),
		joinPartyButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: (height/3)*2 + 100}
			}),
			Text: "Party beitreten",
			Callback: func() {
				client.currentScreen = newJoinPartyScreenLoading(client)
			},
		}),
		changeNameButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
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
	}

	for _, component := range t.components() {
		component.Update()
	}
}

func (t *titleScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)
	for _, component := range t.components() {
		component.Draw(screen)
	}
}

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
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   client.name,
			Colors: &ui.DefaultTitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
	}

	screen.continueButton = ui.NewButton(ui.ButtonConfig{
		Pos: ui.DynamicPosition(func() ui.Position {
			width, height := ebiten.WindowSize()
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
	screen.Fill(ui.DefaultBackgroundColor)

	for _, component := range c.components() {
		component.Draw(screen)
	}
}

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
	if g.client.currentGame != nil {
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
}

func (g *gameScreen) draw(screen *ebiten.Image) {
	if g.client.currentGame != nil {
		g.client.currentGame.Draw(screen)
	}
}

func (g *gameScreen) handlePacket(packet []byte) error {
	if g.client.currentGame != nil {
		err := g.client.currentGame.HandlePacket(packet)
		if err != nil {
			return fmt.Errorf("game failed to handle packet: %w", err)
		}
	}

	return nil
}

type joinPartyScreenLoading struct {
	client      *client
	loadingText *ui.Text
}

var _ packetHandlerScreen = (*joinPartyScreenLoading)(nil)

func newJoinPartyScreenLoading(client *client) *joinPartyScreenLoading {
	queryParties, err := json.Marshal(protocol.QueryPartiesPacket{
		PacketName: protocol.QueryPartiesPacketName,
	})
	if err != nil {
		panic(err)
	}
	client.SendPacket(queryParties)

	return &joinPartyScreenLoading{
		client: client,
		loadingText: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 2}
			}),
			Text: "Lade Partys...",
		}),
	}
}

func (j *joinPartyScreenLoading) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		j.client.currentScreen = newTitleScreen(j.client)
	}

	j.loadingText.Update()
}

func (j *joinPartyScreenLoading) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)
	j.loadingText.Draw(screen)
}

func (j *joinPartyScreenLoading) handlePacket(data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		j.client.currentScreen = newJoinPartyScreenFailed(j.client)
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	switch packetName {
	case protocol.ListPartiesPacketName:
		var listParties protocol.ListPartiesPacket
		err := json.Unmarshal(data, &listParties)
		if err != nil {
			j.client.currentScreen = newJoinPartyScreenFailed(j.client)
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		j.client.currentScreen = newJoinPartyScreenSuccess(j.client, listParties)

		return nil
	default:
		return fmt.Errorf("unknown packet name: %s", packetName)
	}
}

type joinPartyScreenFailed struct {
	client     *client
	failedText *ui.Text
}

func newJoinPartyScreenFailed(client *client) *joinPartyScreenFailed {
	return &joinPartyScreenFailed{
		client: client,
		failedText: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 2}
			}),
			Text: "Fehler beim Laden der Partys",
		}),
	}
}

var _ screen = (*joinPartyScreenFailed)(nil)

func (j *joinPartyScreenFailed) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		j.client.currentScreen = newTitleScreen(j.client)
	}

	j.failedText.Update()
}

func (j *joinPartyScreenFailed) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)
	j.failedText.Draw(screen)
}

type joinPartyScreenSuccess struct {
	client  *client
	title   *ui.Text
	buttons []*ui.Button
}

var _ screen = (*joinPartyScreenSuccess)(nil)

func newJoinPartyScreenSuccess(client *client, packet protocol.ListPartiesPacket) *joinPartyScreenSuccess {
	buttons := make([]*ui.Button, len(packet.Parties))
	for i, party := range packet.Parties {
		iCopy := i
		partyCopy := party

		buttons[i] = ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: (height/3)*2 + 100*iCopy}
			}),
			Text: fmt.Sprintf("%s (%d Spieler)", party.Name, len(party.Players)),
			Callback: func() {
				joinParty, err := json.Marshal(protocol.JoinPartyPacket{
					PacketName: protocol.JoinPartyPacketName,
					Id:         partyCopy.Id,
				})
				if err != nil {
					panic(err)
				}
				client.SendPacket(joinParty)
			},
		})
	}

	return &joinPartyScreenSuccess{
		client: client,
		title: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Party beitreten",
			Colors: &ui.DefaultTitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
		buttons: buttons,
	}
}

func (j *joinPartyScreenSuccess) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		j.client.currentScreen = newTitleScreen(j.client)
	}

	j.title.Update()
	for _, button := range j.buttons {
		button.Update()
	}
}

func (j *joinPartyScreenSuccess) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)

	j.title.Draw(screen)
	for _, button := range j.buttons {
		button.Draw(screen)
	}
}

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
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Name der Party: Neue Party",
			Colors: &ui.DefaultTitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
	}

	screen.continueButton = ui.NewButton(ui.ButtonConfig{
		Pos: ui.DynamicPosition(func() ui.Position {
			width, height := ebiten.WindowSize()
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
	c.partyNameText.Text = c.partyName
	c.partyNameText.Update()
	c.continueButton.Update()
}

func (c *createPartyScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)
	c.partyNameText.Draw(screen)
	c.continueButton.Draw(screen)
}

type partyScreen struct {
	client          *client
	title           *ui.Text
	startGameButton *ui.Button
}

var _ screen = (*partyScreen)(nil)

func newPartyScreen(client *client) *partyScreen {
	return &partyScreen{
		client: client,
		title: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Party: " + client.partyName,
			Colors: &ui.DefaultTitleColors,
			Font:   rescources.RobotoTitleFont,
		}),
		startGameButton: ui.NewButton(ui.ButtonConfig{
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height - 100}
			}),
			Text: "Spiel starten",
			Callback: func() {
				client.currentScreen = newStartGameScreen(client)
			},
		}),
	}
}

func (p *partyScreen) playerList() []*ui.Text {
	width, height := ebiten.WindowSize()

	players := p.client.partyPlayersSorted()

	playerList := make([]*ui.Text, len(players))
	for i, player := range players {
		playerList[i] = ui.NewText(ui.TextConfig{
			Pos:  ui.CenteredPosition{X: width / 2, Y: 100 + height/3 + 100*i},
			Text: player.Name,
		})
	}

	return playerList
}

func (p *partyScreen) components() []ui.Component {
	components := make([]ui.Component, 0)
	components = append(components, p.title, p.startGameButton)

	for _, playerWidget := range p.playerList() {
		components = append(components, playerWidget)
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

	for _, component := range p.components() {
		component.Update()
	}
}

func (p *partyScreen) draw(screen *ebiten.Image) {
	screen.Fill(ui.DefaultBackgroundColor)
	for _, component := range p.components() {
		component.Draw(screen)
	}
}

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
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
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
			Pos: ui.DynamicPosition(func() ui.Position {
				width, height := ebiten.WindowSize()
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Spiel starten",
			Colors: &ui.DefaultTitleColors,
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
	screen.Fill(ui.DefaultBackgroundColor)
	for _, component := range s.components() {
		component.Draw(screen)
	}
}
