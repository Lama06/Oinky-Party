package flappyoinky

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"strconv"
	"time"

	"github.com/Lama06/Oinky-Party/client/game"
	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/Lama06/Oinky-Party/client/ui"
	shared "github.com/Lama06/Oinky-Party/flappyoinky"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
)

var (
	//go:embed oinky.png
	oinkyImageData []byte
	oinkyImage     = loadImage(oinkyImageData)

	//go:embed obstacle.png
	obstacleTileImageData []byte
	obstacleTileImage     = loadImage(obstacleTileImageData)
)

func loadImage(data []byte) image.Image {
	_, img, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("failed to load image: %w", err))
	}
	return img
}

// Die Größe des Oinkys wird anhand der Größe des Fensters zu einem Quadrat ergänzt.
func getVisualOinkySize() int {
	windowWidth, windowHeight := ebiten.WindowSize()
	if windowWidth >= windowHeight {
		return int(shared.OinkySize * float64(windowWidth))
	} else {
		return int(shared.OinkySize * float64(windowHeight))
	}
}

type player struct {
	id           int32   // Die ID des Spielers
	serverPosY   float64 // Die letzte Y Position, die vom Server gesendet wurde
	clientPosY   float64 // Zwischen den UpdatePackets vom Server wird die Y Position der Vögel anhand der vergangenen Zeit des letzten Packets berechnet
	serverSpeedY float64 // Der letzte Speed Wert, der vom Server gesendet wurde
	clientSpeedY float64 // Zwischen den UpdatePackets vom Server wird der Y Speed der Vögel anhand der vergangenen Zeit des letzten Packets berechnet
	rotation     float64 // Die Neigung des Vogels richtet sich nach seiner Geschwindigkeit. Die Veränderung wird gedrosselt, damit eine Animation entsteht
}

func (p *player) clientTick(delta float64) {
	// Y Position und Geschwindigkeit berechnen
	posY := p.serverPosY
	speedY := p.serverSpeedY

	skippedTicks := int(math.Trunc(delta))
	for i := 1; i <= skippedTicks; i++ {
		speedY += shared.OinkyAccelerationY
		posY += speedY
	}
	remainingDelta := delta - float64(skippedTicks)
	posY += speedY * remainingDelta

	p.clientPosY = posY
	p.clientSpeedY = speedY

	// Rotation berechnen
	const maxRotationChangePerTick = 0.05
	targetRotation := p.clientSpeedY * 15
	if p.rotation > targetRotation {
		diff := p.rotation - targetRotation
		if diff > maxRotationChangePerTick {
			diff = maxRotationChangePerTick
		}
		p.rotation -= diff
	} else if p.rotation < targetRotation {
		diff := targetRotation - p.rotation
		if diff > maxRotationChangePerTick {
			diff = maxRotationChangePerTick
		}
		p.rotation += diff
	}
}

func (p *player) draw(client game.Client, screen *ebiten.Image, debug bool) {
	windowWidth, windowHeight := ebiten.WindowSize()
	img := ebiten.NewImageFromImage(oinkyImage)
	imgWidth, imgHeight := img.Size()
	oinkySize := getVisualOinkySize()
	oinkyXScale, oinkyYScale := float64(oinkySize)/float64(imgWidth), float64(oinkySize)/float64(imgHeight)
	oinkyX, oinkyY := shared.OinkyPosX*float64(windowWidth), p.clientPosY*float64(windowHeight)

	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Rotate(p.rotation)
	drawOptions.GeoM.Scale(oinkyXScale, oinkyYScale)
	drawOptions.GeoM.Translate(oinkyX, oinkyY)
	screen.DrawImage(img, &drawOptions)

	if p.id != client.Id() {
		partyPlayer := client.PartyPlayers()[p.id]
		textX, textY := oinkyX+float64(oinkySize)+50, oinkyY+float64(oinkySize)/2
		text.Draw(screen, partyPlayer.Name, rescources.RobotoNormalFont, int(textX), int(textY), colornames.Black)
	}

	if debug && p.id == client.Id() {
		realWidth, realHeight := shared.OinkySize*float64(windowWidth), shared.OinkySize*float64(windowHeight)
		debugImg := ebiten.NewImage(int(realWidth), int(realHeight))
		debugImg.Fill(colornames.Red)

		var drawOptions ebiten.DrawImageOptions
		drawOptions.GeoM.Translate(oinkyX, oinkyY)
		screen.DrawImage(debugImg, &drawOptions)
	}
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

func addObstacleTexture(obstacleImage *ebiten.Image) {
	obstacleImageWidth, obstacleImageHeight := obstacleImage.Size()
	obstacleTileImageEbiten := ebiten.NewImageFromImage(obstacleTileImage)
	obstacleTileImageWidth, obstacleTileImageHeight := obstacleTileImageEbiten.Size()
	obstacleTileImageScaleFactor := float64(obstacleImageWidth) / float64(obstacleTileImageWidth)
	obstacleTileImageScaledHeight := float64(obstacleTileImageHeight) * obstacleTileImageScaleFactor

	for y := 0.0; y < float64(obstacleImageHeight); y += obstacleTileImageScaledHeight {
		var drawOptions ebiten.DrawImageOptions
		drawOptions.GeoM.Scale(obstacleTileImageScaleFactor, obstacleTileImageScaleFactor)
		drawOptions.GeoM.Translate(0, y)
		obstacleImage.DrawImage(obstacleTileImageEbiten, &drawOptions)
	}
}

func (o *obstacle) draw(screen *ebiten.Image) {
	windowWidth, windowHeight := ebiten.WindowSize()
	width := int(shared.ObstacleWidth * float64(windowWidth))

	upperHeight := int(o.freeSpaceUpperY * float64(windowHeight))
	if upperHeight > 0 {
		upper := ebiten.NewImage(width, upperHeight)
		addObstacleTexture(upper)
		var upperDrawOptions ebiten.DrawImageOptions
		upperDrawOptions.GeoM.Translate(o.clientPosX*float64(windowWidth), 0)
		screen.DrawImage(upper, &upperDrawOptions)
	}

	lowerHeight := int((1 - o.freeSpaceLowerY) * float64(windowHeight))
	if lowerHeight > 0 {
		lower := ebiten.NewImage(width, lowerHeight)
		addObstacleTexture(lower)
		var lowerDrawOptions ebiten.DrawImageOptions
		lowerDrawOptions.GeoM.Translate(o.clientPosX*float64(windowWidth), float64(windowHeight)-float64(lowerHeight))
		screen.DrawImage(lower, &lowerDrawOptions)
	}
}

type impl struct {
	client           game.Client
	lastTickTime     int64 // Die Zeit in Millisekunden, bei der das letzte Mal die Daten vom Server aktualisiert wurden
	players          map[int32]*player
	obstacles        []*obstacle
	obstacleCount    int32
	debugModeEnabled bool
}

var _ game.Game = (*impl)(nil)

func create(client game.Client) game.Game {
	return &impl{
		client:       client,
		lastTickTime: time.Now().UnixMilli(),
	}
}

var _ game.Creator = create

func (i *impl) HandleGameStarted() {
	partyPlayers := i.client.PartyPlayers()
	i.players = make(map[int32]*player, len(partyPlayers))
	for id := range partyPlayers {
		i.players[id] = &player{
			id:           id,
			serverPosY:   shared.OinkyStartPosY,
			clientPosY:   shared.OinkyStartPosY,
			serverSpeedY: 0,
			clientSpeedY: 0,
			rotation:     0,
		}
	}
}

func (i *impl) HandleGameEnded() {}

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

		oldRotations := make(map[int32]float64, len(i.players))
		for id, player := range i.players {
			oldRotations[id] = player.rotation
		}
		i.players = make(map[int32]*player, len(update.Players))
		for _, playerUpdateData := range update.Players {
			i.players[playerUpdateData.Player] = &player{
				id:           playerUpdateData.Player,
				serverPosY:   playerUpdateData.PositionY,
				clientPosY:   playerUpdateData.PositionY,
				serverSpeedY: playerUpdateData.SpeedY,
				clientSpeedY: playerUpdateData.SpeedY,
				rotation:     oldRotations[playerUpdateData.Player],
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
		i.obstacleCount = update.ObstacleCount

		i.lastTickTime = time.Now().UnixMilli()

		return nil
	default:
		return fmt.Errorf("unknown packet name: %s", packetName)
	}
}

func (i *impl) obstacleCounter() *ui.Text {
	windowWidth, _ := ebiten.WindowSize()

	return ui.NewText(ui.NewCenteredPosition(windowWidth/2, 50), strconv.Itoa(int(i.obstacleCount)), ui.TextColorPalette{
		Color: colornames.Black,
	}, rescources.RobotoTitleFont)
}

func (i *impl) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightblue)

	for id, player := range i.players {
		if id == i.client.Id() {
			continue
		}

		player.draw(i.client, screen, i.debugModeEnabled)
	}

	if i.alive() {
		i.players[i.client.Id()].draw(i.client, screen, i.debugModeEnabled)
	}

	for _, obstacle := range i.obstacles {
		obstacle.draw(screen)
	}

	i.obstacleCounter().Draw(screen)
}

func (i *impl) Update() {
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		i.debugModeEnabled = !i.debugModeEnabled
	}

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

	i.obstacleCounter().Update()
}

func (i *impl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (i *impl) alive() bool {
	_, ok := i.players[i.client.Id()]
	return ok
}

func (i *impl) delta() float64 {
	currentTime := time.Now().UnixMilli()
	deltaTime := float64(currentTime-i.lastTickTime) / protocol.TickSpeed
	return deltaTime
}

var Type = game.Type{
	Creator:     create,
	Name:        shared.Name,
	DisplayName: "Flappy Oinky",
}
