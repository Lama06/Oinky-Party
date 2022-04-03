package flappyoinky

import (
	"encoding/json"
	"errors"
	"fmt"
	shared "github.com/Lama06/Oinky-Party/flappyoinky"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
	"math/rand"
)

type player struct {
	id        int32   // Die ID des Spielers, zu dem dieser Vogel gehört
	positionY float64 // Die Y Position der oberen linken Ecke des Vogels
	speedY    float64 // Die Geschwindigkeit, mit der positionY pro Tick erhöht wird
}

func (p *player) isOutsideWorld() bool {
	return p.positionY <= 0 || p.positionY-shared.OinkySize >= 1
}

func (p *player) isTouchingObstacle(obstacles []*obstacle) bool {
	for _, obstacle := range obstacles {
		if obstacle.posX >= shared.OinkyPosX+shared.OinkySize || obstacle.posX+shared.ObstacleWidth <= shared.OinkyPosX {
			continue
		}

		if p.positionY >= obstacle.freeSpaceUpperY && p.positionY+shared.OinkySize <= obstacle.freeSpaceLowerY {
			continue
		}

		if p.positionY+shared.OinkySize <= obstacle.freeSpaceLowerY && p.positionY >= obstacle.freeSpaceUpperY {
			continue
		}

		return true
	}

	return false
}

func (p *player) tick() {
	p.speedY += shared.OinkySpeedYIncreasePerTick
	p.positionY += p.speedY
}

func (p *player) jump() {
	p.speedY = shared.OinkySpeedYAfterJump
}

func (p *player) toUpdateData() shared.PlayerUpdateData {
	return shared.PlayerUpdateData{
		Player:    p.id,
		PositionY: p.positionY,
		SpeedY:    p.speedY,
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
	alivePlayers           map[int32]*player
	ticksUntilNextObstacle int
	obstacleCount          int32
	obstacles              []*obstacle
}

var _ game.Game = (*impl)(nil)

func Create(party game.Party) game.Game {
	return &impl{
		party:                  party,
		alivePlayers:           make(map[int32]*player, len(party.Players())),
		obstacles:              make([]*obstacle, 0),
		ticksUntilNextObstacle: shared.ObstacleSpawnRate,
	}
}

var _ game.Creator = Create

func (i *impl) HandleGameStarted() {
	for _, partyPlayer := range i.party.Players() {
		i.alivePlayers[partyPlayer.Id()] = &player{
			id:        partyPlayer.Id(),
			positionY: shared.OinkyStartPosY,
			speedY:    0,
		}
	}
}

func (i *impl) HandleGameEnded() {}

func (i *impl) HandlePlayerLeft(player game.Player) {
	delete(i.alivePlayers, player.Id())

	if len(i.alivePlayers) == 0 {
		i.party.EndGame()
	}
}

func (i *impl) HandlePacket(sender game.Player, data []byte) error {
	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to obtain the packet name: %w", err)
	}

	switch packetName {
	case shared.JumpPacketName:
		data, ok := i.alivePlayers[sender.Id()]
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
	players := make([]shared.PlayerUpdateData, 0, len(i.alivePlayers))
	for _, player := range i.alivePlayers {
		players = append(players, player.toUpdateData())
	}

	obstacles := make([]shared.ObstacleUpdateData, len(i.obstacles))
	for i, obstacle := range i.obstacles {
		obstacles[i] = obstacle.toUpdateData()
	}

	update, err := json.Marshal(shared.UpdatePacket{
		PacketName:    shared.UpdatePacketName,
		Players:       players,
		Obstacles:     obstacles,
		ObstacleCount: i.obstacleCount,
	})
	if err != nil {
		panic(err)
	}
	i.party.BroadcastPacket(update)
}

func (i *impl) tickPlayers() (gameEnded bool) {
	for id, player := range i.alivePlayers {
		partyPlayer := i.party.Players()[id]

		player.tick()

		if player.isOutsideWorld() || player.isTouchingObstacle(i.obstacles) {
			gameEnded := i.killPlayer(partyPlayer)
			if gameEnded {
				return true
			}
		}
	}

	return false
}

func (i *impl) killPlayer(player game.Player) (gameEnded bool) {
	delete(i.alivePlayers, player.Id())

	if len(i.alivePlayers) == 0 {
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
	i.obstacleCount++

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
