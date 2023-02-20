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
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
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
	screen.Fill(ui.BackgroundColor)
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

var _ screen = (*joinPartyScreenFailed)(nil)

func newJoinPartyScreenFailed(client *client) *joinPartyScreenFailed {
	return &joinPartyScreenFailed{
		client: client,
		failedText: ui.NewText(ui.TextConfig{
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 2}
			}),
			Text: "Fehler beim Laden der Partys",
		}),
	}
}

func (j *joinPartyScreenFailed) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		j.client.currentScreen = newTitleScreen(j.client)
	}

	j.failedText.Update()
}

func (j *joinPartyScreenFailed) draw(screen *ebiten.Image) {
	screen.Fill(ui.BackgroundColor)
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
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height/3*2 + 100*iCopy}
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
			Pos: ui.DynamicPosition(func(width, height int) ui.Position {
				return ui.CenteredPosition{X: width / 2, Y: height / 3}
			}),
			Text:   "Party beitreten",
			Colors: &ui.TitleColors,
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
	screen.Fill(ui.BackgroundColor)

	j.title.Draw(screen)
	for _, button := range j.buttons {
		button.Draw(screen)
	}
}
