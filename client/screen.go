package client

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	defaultBackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	defaultButtonColors    = ui.ButtonColorPalette{
		BackgroundColor:      color.RGBA{R: 18, G: 53, B: 91, A: 255},
		BackgroundHoverColor: color.RGBA{R: 134, G: 22, B: 87, A: 255},
		TextColor:            color.RGBA{R: 212, G: 245, B: 245, A: 255},
		TextHoverColor:       color.RGBA{R: 212, G: 245, B: 245, A: 255},
	}
	defaultTextColors = ui.TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 82, G: 73, B: 72, A: 255},
	}
	defaultTitleColors = ui.TextColorPalette{
		Color:      color.RGBA{R: 87, G: 70, B: 123, A: 255},
		HoverColor: color.RGBA{R: 112, G: 248, B: 186, A: 255},
	}
)

type screen interface {
	Update()
	Draw(screen *ebiten.Image)
}

type packetHandlerScreen interface {
	screen
	HandlePacket(packet []byte) error
}

type titleScreen struct {
	c *client
}

var _ screen = (*titleScreen)(nil)

func newTitleScreen(c *client) *titleScreen {
	return &titleScreen{
		c: c,
	}
}

func (t *titleScreen) title() *ui.Text {
	width, height := ebiten.WindowSize()

	return ui.NewText(ui.NewCenteredPosition(width/2, height/3), "Oinky Party", defaultTitleColors, rescources.RobotoTitleFont)
}

func (t *titleScreen) createPartyButton() *ui.Button {
	width, height := ebiten.WindowSize()

	return ui.NewButton(ui.NewCenteredPosition(width/2, (height/3)*2), "Party erstellen", defaultButtonColors, func() {
		t.c.currentScreen = newCreatePartyScreen(t.c)
	})
}

func (t *titleScreen) joinPartyButton() *ui.Button {
	width, height := ebiten.WindowSize()

	return ui.NewButton(ui.NewCenteredPosition(width/2, (height/3)*2+100), "Party beitreten", defaultButtonColors, func() {
		t.c.currentScreen = newJoinPartyScreen(t.c)
	})
}

func (t *titleScreen) changeNameButton() *ui.Button {
	width, height := ebiten.WindowSize()

	return ui.NewButton(ui.NewCenteredPosition(width/2, (height/3)*2+200), "Namen ändern", defaultButtonColors, func() {
		t.c.currentScreen = newChangeNameScreen(t.c)
	})
}

func (t *titleScreen) content() []ui.Component {
	return []ui.Component{t.title(), t.createPartyButton(), t.joinPartyButton(), t.changeNameButton()}
}

func (t *titleScreen) Update() {
	if inpututil.IsKeyJustReleased(ebiten.Key1) {
		t.c.currentScreen = newCreatePartyScreen(t.c)
	} else if inpututil.IsKeyJustReleased(ebiten.Key2) {
		t.c.currentScreen = newJoinPartyScreen(t.c)
	}

	for _, component := range t.content() {
		component.Update()
	}
}

func (t *titleScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	for _, component := range t.content() {
		component.Draw(screen)
	}
}

type changeNameScreen struct {
	c       *client
	newName []rune
}

var _ screen = (*changeNameScreen)(nil)

func newChangeNameScreen(c *client) *changeNameScreen {
	return &changeNameScreen{
		c:       c,
		newName: []rune(c.name),
	}
}

func (c *changeNameScreen) nameText() *ui.Text {
	windowWidth, windowHeight := ebiten.WindowSize()
	return ui.NewText(ui.NewCenteredPosition(windowWidth/2, windowHeight/3), string(c.newName), defaultTitleColors, rescources.RobotoTitleFont)
}

func (c *changeNameScreen) changeButton() *ui.Button {
	windowWidth, windowHeight := ebiten.WindowSize()
	return ui.NewButton(ui.NewCenteredPosition(windowWidth/2, (windowHeight/3)*2), "Namen ändern", defaultButtonColors, func() {
		changeName, err := json.Marshal(protocol.ChangeNamePacket{
			PacketName: protocol.ChangeNamePacketName,
			NewName:    string(c.newName),
		})
		if err != nil {
			panic(err)
		}
		c.c.SendPacket(changeName)

		c.c.name = string(c.newName)

		c.c.currentScreen = newTitleScreen(c.c)
	})
}

func (c *changeNameScreen) content() []ui.Component {
	return []ui.Component{c.nameText(), c.changeButton()}
}

func (c *changeNameScreen) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		c.c.currentScreen = newTitleScreen(c.c)
	}

	c.newName = ebiten.AppendInputChars(c.newName)
	if len(c.newName) != 0 && inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		c.newName = c.newName[:len(c.newName)-1]
	}

	for _, component := range c.content() {
		component.Update()
	}
}

func (c *changeNameScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)

	for _, component := range c.content() {
		component.Draw(screen)
	}
}

type gameScreen struct {
	c *client
}

var _ packetHandlerScreen = (*gameScreen)(nil)

func newGameScreen(c *client) *gameScreen {
	return &gameScreen{
		c: c,
	}
}

func (s *gameScreen) Update() {
	if s.c.currentGame != nil {
		s.c.currentGame.Update()

		if inpututil.IsKeyJustReleased(ebiten.KeyEscape) {
			endGame, err := json.Marshal(protocol.EndGamePacket{
				PacketName: protocol.EndGamePacketName,
			})
			if err != nil {
				panic(err)
			}
			s.c.SendPacket(endGame)
		}
	}
}

func (s *gameScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	if s.c.currentGame != nil {
		s.c.currentGame.Draw(screen)
	}
}

func (s *gameScreen) HandlePacket(packet []byte) error {
	if s.c.currentGame != nil {
		err := s.c.currentGame.HandlePacket(packet)
		if err != nil {
			return fmt.Errorf("game failed to handle packet: %w", err)
		}
	}

	return nil
}

type joinPartyScreen struct {
	c       *client
	failed  bool
	parties []protocol.PartyData
}

var _ packetHandlerScreen = (*joinPartyScreen)(nil)

func newJoinPartyScreen(c *client) *joinPartyScreen {
	queryParties, err := json.Marshal(protocol.QueryPartiesPacket{
		PacketName: protocol.QueryPartiesPacketName,
	})
	if err != nil {
		panic(err)
	}
	c.SendPacket(queryParties)

	return &joinPartyScreen{
		c: c,
	}
}

func (j *joinPartyScreen) statusText() *ui.Text {
	windowWidth, windowHeight := ebiten.WindowSize()

	text := "Lade Parties..."
	if j.failed {
		text = "Fehler beim Laden der Parties"
	}

	return ui.NewText(ui.NewCenteredPosition(windowWidth/2, windowHeight/2), text, defaultTextColors, rescources.RobotoNormalFont)
}

func (j *joinPartyScreen) title() *ui.Text {
	windowWidth, windowHeight := ebiten.WindowSize()

	return ui.NewText(ui.NewCenteredPosition(windowWidth/2, windowHeight/3), "Party beitreten", defaultTitleColors, rescources.RobotoTitleFont)
}

func (j *joinPartyScreen) partiesList() []*ui.Button {
	windowWidth, windowHeight := ebiten.WindowSize()

	partiesList := make([]*ui.Button, 0)
	for i, party := range j.parties {
		partyCopy := party

		partyButton := ui.NewButton(ui.NewCenteredPosition(
			windowWidth/2,
			(windowHeight/3)*2+100*i,
		), fmt.Sprintf("%s (%d Spieler)", party.Name, len(party.Players)), defaultButtonColors, func() {
			joinParty, err := json.Marshal(protocol.JoinPartyPacket{
				PacketName: protocol.JoinPartyPacketName,
				Id:         partyCopy.Id,
			})
			if err != nil {
				panic(err)
			}
			j.c.SendPacket(joinParty)
		})
		partiesList = append(partiesList, partyButton)
	}

	return partiesList
}

func (j *joinPartyScreen) content() []ui.Component {
	if j.failed || j.parties == nil {
		return []ui.Component{j.statusText()}
	} else {
		partyButtons := j.partiesList()
		components := []ui.Component{j.title()}
		for _, partyButton := range partyButtons {
			components = append(components, partyButton)
		}
		return components
	}
}

func (j *joinPartyScreen) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		j.c.currentScreen = newTitleScreen(j.c)
	}

	for _, component := range j.content() {
		component.Update()
	}
}

func (j *joinPartyScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	for _, component := range j.content() {
		component.Draw(screen)
	}
}

func (j *joinPartyScreen) HandlePacket(packet []byte) error {
	packetName, err := protocol.GetPacketName(packet)
	if err != nil {
		j.failed = true
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	if packetName == protocol.ListPartiesPacketName {
		var listParties protocol.ListPartiesPacket
		err := json.Unmarshal(packet, &listParties)
		if err != nil {
			j.failed = true
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		j.failed = false
		j.parties = listParties.Parties
	}

	return nil
}

type createPartyScreen struct {
	c         *client
	partyName []rune
}

var _ screen = (*createPartyScreen)(nil)

func newCreatePartyScreen(c *client) *createPartyScreen {
	return &createPartyScreen{
		c:         c,
		partyName: []rune("Neue Party"),
	}
}

func (c *createPartyScreen) partyNameText() *ui.Text {
	width, height := ebiten.WindowSize()

	pos := ui.NewCenteredPosition(width/2, height/3)

	return ui.NewText(pos, "Name der Party: "+string(c.partyName), defaultTitleColors, rescources.RobotoTitleFont)
}

func (c *createPartyScreen) createButton() *ui.Button {
	width, height := ebiten.WindowSize()

	pos := ui.NewCenteredPosition(width/2, height/3*2)

	callback := func() {
		createParty, err := json.Marshal(protocol.CreatePartyPacket{
			PacketName: protocol.CreatePartyPacketName,
			Name:       string(c.partyName),
		})
		if err != nil {
			panic(err)
		}
		c.c.SendPacket(createParty)
	}

	return ui.NewButton(pos, "Party erstellen", defaultButtonColors, callback)
}

func (c *createPartyScreen) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		c.c.currentScreen = newTitleScreen(c.c)
	}

	c.partyName = ebiten.AppendInputChars(c.partyName)
	if len(c.partyName) != 0 && inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		c.partyName = c.partyName[:len(c.partyName)-1]
	}
	c.partyNameText().Update()
	c.createButton().Update()
}

func (c *createPartyScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	c.partyNameText().Draw(screen)
	c.createButton().Draw(screen)
}

type partyScreen struct {
	c *client
}

var _ screen = (*partyScreen)(nil)

func newPartyScreen(c *client) *partyScreen {
	return &partyScreen{
		c: c,
	}
}

func (p *partyScreen) title() *ui.Text {
	width, height := ebiten.WindowSize()

	pos := ui.NewCenteredPosition(width/2, height/3)

	return ui.NewText(pos, "Party: "+p.c.partyName, defaultTitleColors, rescources.RobotoTitleFont)
}

func (p *partyScreen) playerList() []*ui.Text {
	windowWidth, windowHeight := ebiten.WindowSize()

	playerList := make([]*ui.Text, 0)
	i := 0
	for _, player := range p.c.partyPlayersSorted() {
		playerList = append(playerList, ui.NewText(ui.NewCenteredPosition(
			windowWidth/2,
			100+windowHeight/3+100*i,
		), player.Name, defaultTextColors, rescources.RobotoNormalFont))
		i++
	}

	return playerList
}

func (p *partyScreen) startGameButton() *ui.Button {
	windowWidth, windowHeight := ebiten.WindowSize()

	pos := ui.NewCenteredPosition(windowWidth/2, windowHeight-100)

	return ui.NewButton(pos, "Spiel starten", defaultButtonColors, func() {
		p.c.currentScreen = newStartGameScreen(p.c)
	})
}

func (p *partyScreen) contents() []ui.Component {
	components := make([]ui.Component, 0)
	components = append(components, p.title(), p.startGameButton())

	for _, playerWidget := range p.playerList() {
		components = append(components, playerWidget)
	}

	return components
}

func (p *partyScreen) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		leaveParty, err := json.Marshal(protocol.LeavePartyPacket{
			PacketName: protocol.LeavePartyPacketName,
		})
		if err != nil {
			panic(err)
		}
		p.c.SendPacket(leaveParty)
	}

	for _, component := range p.contents() {
		component.Update()
	}
}

func (p *partyScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	for _, component := range p.contents() {
		component.Draw(screen)
	}
}

type startGameScreen struct {
	c *client
}

var _ screen = (*startGameScreen)(nil)

func newStartGameScreen(c *client) *startGameScreen {
	return &startGameScreen{
		c: c,
	}
}

func (s *startGameScreen) title() *ui.Text {
	width, height := ebiten.WindowSize()

	pos := ui.NewCenteredPosition(width/2, height/3)

	return ui.NewText(pos, "Spiel starten", defaultTextColors, rescources.RobotoTitleFont)
}

func (s *startGameScreen) gameButtons() []*ui.Button {
	windowWidth, windowHeight := ebiten.WindowSize()

	buttons := make([]*ui.Button, 0)
	for i, gameType := range gameTypes {
		pos := ui.NewCenteredPosition(windowWidth/2, 100+windowHeight/3+100*i)

		gameTypeCopy := gameType

		callback := func() {
			startGame, err := json.Marshal(protocol.StartGamePacket{
				PacketName: protocol.StartGamePacketName,
				GameType:   gameTypeCopy.Name,
			})
			if err != nil {
				panic(err)
			}
			s.c.SendPacket(startGame)
		}

		buttons = append(buttons, ui.NewButton(pos, gameType.DisplayName, defaultButtonColors, callback))
	}

	return buttons
}

func (s *startGameScreen) content() []ui.Component {
	components := make([]ui.Component, 0)
	components = append(components, s.title())

	for _, button := range s.gameButtons() {
		components = append(components, button)
	}

	return components
}

func (s *startGameScreen) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.c.currentScreen = newPartyScreen(s.c)
	}

	for _, component := range s.content() {
		component.Update()
	}
}

func (s *startGameScreen) Draw(screen *ebiten.Image) {
	screen.Fill(defaultBackgroundColor)
	for _, component := range s.content() {
		component.Draw(screen)
	}
}
