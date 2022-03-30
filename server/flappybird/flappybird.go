package flappybird

import (
	"encoding/json"
	"errors"
	"fmt"
	shared "github.com/Lama06/Oinky-Party/flappybird"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
	"math/rand"
	"time"
)

type bird struct {
	id        int32   // Die ID des Spielers, zu dem dieser Vogel gehört
	positionY float64 // Die Y Position der oberen linken Ecke des Vogels
	speedY    float64 // Die Geschwindigkeit, mit der positionY pro Tick erhöht wird
}

func (b *bird) isOutsideWorld() bool {
	return b.positionY <= 0 || b.positionY-shared.BirdSize >= 1
}

func (b *bird) isTouchingObstacle(obstacles []*obstacle) bool {
	for _, obstacle := range obstacles {
		if obstacle.posX >= shared.BirdPosX+shared.BirdSize || obstacle.posX+shared.ObstacleWidth <= shared.BirdPosX {
			continue
		}

		if b.positionY >= obstacle.freeSpaceUpperY && b.positionY+shared.BirdSize <= obstacle.freeSpaceLowerY {
			continue
		}

		if b.positionY+shared.BirdSize <= obstacle.freeSpaceLowerY && b.positionY >= obstacle.freeSpaceUpperY {
			continue
		}

		return true
	}

	return false
}

func (b *bird) tick() {
	b.speedY += shared.BirdSpeedYIncreasePerTick
	b.positionY += b.speedY
}

func (b *bird) jump() {
	b.speedY = shared.BirdSpeedYAfterJump
}

func (b *bird) toUpdateData() shared.PlayerUpdateData {
	return shared.PlayerUpdateData{
		Player:    b.id,
		PositionY: b.positionY,
		SpeedY:    b.speedY,
	}
}

func randomObstacleFreeSpace() (lowerY, upperY float64) {
	freeSpaceLowerY := 1 - float64(rand.Intn(10-shared.ObstacleFreeSpaceHeight*10))*0.1
	freeSpaceUpperY := freeSpaceLowerY - shared.ObstacleFreeSpaceHeight
	return freeSpaceLowerY, freeSpaceUpperY
}

type obstacle struct {
	freeSpaceLowerY float64 // Die Y Koordinate, die die untere Begrenzung des freien Platzes angibt
	freeSpaceUpperY float64 // Die Y Koordinate, die die obere Begrenzung des freien Platzes angibt
	posX            float64 // Die X Position der linken Kante des Hindernisses
}

func (o *obstacle) isOutsideWorld() bool {
	return o.posX+shared.ObstacleWidth < 0
}

func (o *obstacle) tick() {
	o.posX += shared.ObstacleSpeed
}

func (o obstacle) toUpdateData() shared.ObstacleUpdateData {
	return shared.ObstacleUpdateData{
		FreeSpaceLowerY: o.freeSpaceLowerY,
		FreeSpaceUpperY: o.freeSpaceUpperY,
		PosX:            o.posX,
	}
}

type impl struct {
	party                  game.Party
	playerDataMap          map[int32]*bird
	ticksUntilNextObstacle int
	obstacles              []*obstacle
}

var _ game.Game = (*impl)(nil)

func Create(party game.Party) game.Game {
	return &impl{
		party:                  party,
		playerDataMap:          make(map[int32]*bird, len(party.Players())),
		obstacles:              make([]*obstacle, 0),
		ticksUntilNextObstacle: shared.ObstacleSpawnRate,
	}
}

var _ game.Creator = Create

func (i *impl) HandleGameStarted() {
	for _, player := range i.party.Players() {
		i.playerDataMap[player.Id()] = &bird{
			id:        player.Id(),
			positionY: shared.BirdStartPosY,
			speedY:    0,
		}
	}
}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePlayerLeft(player game.Player) {
	delete(i.playerDataMap, player.Id())
}

func (i *impl) HandlePacket(sender game.Player, data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to obtain the packet name: %w", err)
	}

	switch packetName {
	case shared.JumpPacketName:
		data, ok := i.playerDataMap[sender.Id()]
		if !ok {
			return errors.New("invalid player id")
		}

		data.jump()
	}

	return nil
}

func (i *impl) Tick() {
	gameEnded := i.tickPlayers()
	if gameEnded {
		return
	}
	i.tickObstacles()
	i.broadcastUpdatePacket()
}

func (i *impl) broadcastUpdatePacket() {
	players := make([]shared.PlayerUpdateData, 0, len(i.playerDataMap))
	for _, data := range i.playerDataMap {
		players = append(players, data.toUpdateData())
	}

	obstacles := make([]shared.ObstacleUpdateData, len(i.obstacles))
	for i, obstacle := range i.obstacles {
		obstacles[i] = obstacle.toUpdateData()
	}

	update, err := json.Marshal(shared.UpdatePacket{
		PacketName: shared.UpdatePacketName,
		Time: time.Now().UnixMilli(),
		Players:    players,
		Obstacles:  obstacles,
	})
	if err != nil {
		panic(err)
	}
	i.party.BroadcastPacket(update)
}

func (i *impl) tickPlayers() (gameEnded bool) {
	for id, data := range i.playerDataMap {
		player := i.party.Server().PlayerById(id)

		data.tick()

		if data.isOutsideWorld() || data.isTouchingObstacle(i.obstacles) {
			gameEnded := i.killPlayer(player)
			if gameEnded {
				return true
			}
		}
	}

	return false
}

func (i *impl) killPlayer(player game.Player) (gameEnded bool) {
	delete(i.playerDataMap, player.Id())

	if len(i.playerDataMap) == 0 {
		i.party.EndGame()
		return true
	}

	return false
}

func (i *impl) tickObstacles() {
	i.ticksUntilNextObstacle--

	if i.ticksUntilNextObstacle == 0 {
		i.ticksUntilNextObstacle = shared.ObstacleSpawnRate

		i.spawnNewObstacle()
	}

	i.filterObstacles()

	for _, obstacle := range i.obstacles {
		obstacle.tick()
	}
}

func (i *impl) spawnNewObstacle() {
	freeSpaceLowerY, freeSpaceUpperY := randomObstacleFreeSpace()

	newObstacle := &obstacle{
		freeSpaceLowerY: freeSpaceLowerY,
		freeSpaceUpperY: freeSpaceUpperY,
		posX:            1,
	}

	i.obstacles = append(i.obstacles, newObstacle)
}

func (i *impl) filterObstacles() {
	obstacles := make([]*obstacle, 0, len(i.obstacles))
	for _, o := range i.obstacles {
		if !o.isOutsideWorld() {
			obstacles = append(obstacles, o)
		}
	}
	i.obstacles = obstacles
}
