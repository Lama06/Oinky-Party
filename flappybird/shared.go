package flappybird

// Die X und Y Koordinaten der Vögel und Hindernisse sind vom Typ float64 und liegen im Bereich 0 bis 1.
// Der Punkt (0, 0) liegt in der oberen linken Ecke des Bildschirmes.
// Die Koordinaten von den Vögeln und Hindernissen geben Auskunft über die Position der oberen linken Ecke der jeweiligen Objekte.
// Der X/Y Speed Wert der Vögel und Hindernisse gibt an, um wie viel die X/Y Position des Objektes pro Tick erhöht wird.
// Der Speed Wert der Vögel steigt pro Tick, während der, der Hindernisse konstant ist.
// Jeden Tick wird ein UpdatePacket vom Server an alle Spieler versand.
// Die Spieler teilen dem Server mit, wenn ihr eigener Vogel springen soll.
// Der Server entscheidet darüber, ob Spieler gestorben sind.
// Wenn ein Spieler gestorben ist, erkennt der Client das daran, dass dieser Spieler nicht mehr im UpdatePacket zu finden ist.

const (
	BirdSize                  = 0.06             // Die Höhe und Breite des Vogels
	BirdPosX                  = 0.5 - BirdSize/2 // Die permanente X Position der oberen linken Ecke des Vogels
	BirdStartPosY             = 0.5 - BirdSize/2 // Die Y Position der oberen linken Ecke der Vögel, bei der sie sich am Anfang des Spieles befinden
	BirdSpeedYIncreasePerTick = 0.001            // Die
	BirdSpeedYAfterJump       = -0.02

	ObstacleSpawnRate       = 70
	ObstacleWidth           = 0.02
	ObstacleFreeSpaceHeight = 0.3
	ObstacleSpeed           = -0.005
)

// Client zu Server

const JumpPacketName = "flappy-bird-jump"

type JumpPacket struct {
	PacketName string
}

// Server zu Client

type PlayerUpdateData struct {
	Player    int32   // Die ID des Spielers
	PositionY float64 // Die aktuelle Y Position des Vogels
	SpeedY    float64 // Die Geschwindigkeit, mit der PositionY pro Tick erhöht wird
}

type ObstacleUpdateData struct {
	FreeSpaceLowerY float64 // Die Y Koordinate der oberen Kante des unteren Teils des Hindernisses
	FreeSpaceUpperY float64 // Die Y Koordinate der unteren Kante des oberen Teils des Hindernisses
	PosX            float64 // Die aktuelle X Position der linken Kante des Hindernisses
}

const UpdatePacketName = "flappy-bird-update"

type UpdatePacket struct {
	PacketName string
	Time       int64
	Players    []PlayerUpdateData
	Obstacles  []ObstacleUpdateData
}
