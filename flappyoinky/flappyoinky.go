package flappyoinky

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
	OinkySize                  = 0.06              // Die Höhe und Breite des Vogels
	OinkyPosX                  = 0.5 - OinkySize/2 // Die permanente X Position der oberen linken Ecke des Vogels
	OinkyStartPosY             = 0.5 - OinkySize/2 // Die Y Position der oberen linken Ecke der Vögel, bei der sie sich am Anfang des Spieles befinden
	OinkySpeedYIncreasePerTick = 0.001             // Der Wert, mit dem die Geschwindigkeit der Vögel jede Sekunde erhöht wird
	OinkySpeedYAfterJump       = -0.02             // Der Wert der Geschwindigkeit der Vögel nach einem Sprung

	ObstacleSpawnRate       = 70     // Der Abstand in Ticks, in dem Hindernisse spawnen
	ObstacleWidth           = 0.06   // Die Breite der Hindernisse
	ObstacleFreeSpaceHeight = 0.4    // Die Höhe des freien Platzes der Hindernisse
	ObstacleSpeed           = -0.005 // Die Geschwindigkeit, mit der die X Koordinate der Hindernisse pro Tick erhöht wird
)

// Client zu Server

const JumpPacketName = "oinky-bird-jump"

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

const UpdatePacketName = "oinky-bird-update"

type UpdatePacket struct {
	PacketName    string
	Players       []PlayerUpdateData
	Obstacles     []ObstacleUpdateData
	ObstacleCount int32
}
