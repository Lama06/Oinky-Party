package flappybird

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/Lama06/Oinky-Party/client/game"
	shared "github.com/Lama06/Oinky-Party/flappybird"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
	"image"
	_ "image/png"
	"time"
)

var (
	//go:embed bird.png
	birdImageData []byte
	birdImage     = loadImage(birdImageData)
)

func loadImage(data []byte) image.Image {
	_, img, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("failed to load image: %w", err))
	}
	return img
}

// Die Größe des Vogels wird anhand der Größe des Fensters zu einem Quadrat ergänzt.
func getVisualBirdSize() int {
	windowWidth, windowHeight := ebiten.WindowSize()
	if windowWidth >= windowHeight {
		return int(shared.BirdSize * float64(windowWidth))
	} else {
		return int(shared.BirdSize * float64(windowHeight))
	}
}

type player struct {
	id           int32   // Die ID des Spielers
	serverPosY   float64 // Die letzte Y Position, die vom Server gesendet wurde
	clientPosY   float64 // Zwischen den UpdatePackets vom Server wird die Y Position der Vögel anhand der vergangenen Zeit des letzten Packets berechnet
	serverSpeedY float64 // Der letzte Speed Wert, der vom Server gesendet wurde
}

func (p *player) clientTick(delta float64) {
	p.clientPosY = p.serverPosY + p.serverSpeedY*delta
}

func (p player) draw(client game.Client, screen *ebiten.Image) {
	windowWidth, windowHeight := ebiten.WindowSize()
	img := ebiten.NewImageFromImage(birdImage)
	imgWidth, imgHeight := img.Size()
	birdSize := getVisualBirdSize()
	birdXScale, birdYScale := float64(birdSize)/float64(imgWidth), float64(birdSize)/float64(imgHeight)

	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Scale(birdXScale, birdYScale)
	drawOptions.GeoM.Translate(shared.BirdPosX*float64(windowWidth), p.clientPosY*float64(windowHeight))
	screen.DrawImage(img, &drawOptions)
}

type obstacle struct {
	freeSpaceLowerY float64
	freeSpaceUpperY float64
	serverPosX      float64
	clientPosX      float64
}

func (o *obstacle) clientTick(delta float64) {
	o.clientPosX = o.serverPosX + shared.ObstacleSpeed*delta
}

func (o obstacle) draw(screen *ebiten.Image) {
	windowWidth, windowHeight := ebiten.WindowSize()
	width := int(shared.ObstacleWidth * float64(windowWidth))

	upper := ebiten.NewImage(width, int(o.freeSpaceUpperY*float64(windowHeight)))
	upper.Fill(colornames.White)
	var upperDrawOptions ebiten.DrawImageOptions
	upperDrawOptions.GeoM.Translate(o.clientPosX*float64(windowWidth), 0)
	screen.DrawImage(upper, &upperDrawOptions)

	lowerHeight := int((1 - o.freeSpaceLowerY) * float64(windowHeight))
	if lowerHeight > 0 { // Wenn lowerHeight 0 ist erzeugt der Aufruf von NewImage einen panic
		lower := ebiten.NewImage(width, lowerHeight)
		lower.Fill(colornames.White)
		var lowerDrawOptions ebiten.DrawImageOptions
		lowerDrawOptions.GeoM.Translate(o.clientPosX*float64(windowWidth), float64(windowHeight)-float64(lowerHeight))
		screen.DrawImage(lower, &lowerDrawOptions)
	}
}

type impl struct {
	client       game.Client
	lastTickTime int64 // Die Zeit in Millisekunden, bei der das letzte Mal die Daten vom Server aktualisiert wurden
	players      []*player
	obstacles    []*obstacle
}

var _ game.Game = (*impl)(nil)

func Create(client game.Client) game.Game {
	return &impl{
		client: client,
	}
}

var _ game.Creator = Create

func (i *impl) HandleGameStarted() {
	i.lastTickTime = time.Now().UnixMilli()

	partyPlayers := i.client.PartyPlayers()
	i.players = make([]*player, len(partyPlayers))
	for index, partyPlayer := range partyPlayers {
		i.players[index] = &player{
			id:           partyPlayer.Id,
			serverPosY:   shared.BirdStartPosY,
			clientPosY:   shared.BirdStartPosY,
			serverSpeedY: 0,
		}
	}

	i.obstacles = nil
}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePlayerLeft() {}

func (i *impl) HandlePacket(packet []byte) error {
	packetName, err := protocol.GetPacketName(packet)
	if err != nil {
		return fmt.Errorf("could not get packet name: %w", err)
	}

	switch packetName {
	case shared.UpdatePacketName:
		var update shared.UpdatePacket
		err := json.Unmarshal(packet, &update)
		if err != nil {
			return fmt.Errorf("could not unmarshal packet: %w", err)
		}

		i.players = make([]*player, len(update.Players))
		for index, playerData := range update.Players {
			i.players[index] = &player{
				id:           playerData.Player,
				serverPosY:   playerData.PositionY,
				clientPosY:   playerData.PositionY,
				serverSpeedY: playerData.SpeedY,
			}
		}

		i.obstacles = make([]*obstacle, len(update.Obstacles))
		for index, obstacleData := range update.Obstacles {
			i.obstacles[index] = &obstacle{
				freeSpaceLowerY: obstacleData.FreeSpaceLowerY,
				freeSpaceUpperY: obstacleData.FreeSpaceUpperY,
				serverPosX:      obstacleData.PosX,
				clientPosX:      obstacleData.PosX,
			}
		}

		i.lastTickTime = update.Time
	}

	return nil
}

func (i *impl) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Black)

	for _, player := range i.players {
		player.draw(i.client, screen)
	}

	for _, obstacle := range i.obstacles {
		obstacle.draw(screen)
	}
}

func (i *impl) Update() {
	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)) && i.alive() {
		jump, err := json.Marshal(shared.JumpPacket{
			PacketName: shared.JumpPacketName,
		})
		if err != nil {
			panic(err)
		}
		i.client.SendPacket(jump)
	}

	delta := i.delta()

	for _, obstacle := range i.obstacles {
		obstacle.clientTick(delta)
	}

	for _, player := range i.players {
		player.clientTick(delta)
	}

}

func (i *impl) alive() bool {
	for _, player := range i.players {
		if player.id == i.client.Id() {
			return true
		}
	}

	return false
}

func (i *impl) delta() float64 {
	currentTime := time.Now().UnixMilli()
	deltaTime := float64(currentTime-i.lastTickTime) / protocol.TickSpeed
	return deltaTime
}
