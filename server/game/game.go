package game

type Creator func(party Party) Game

type Player interface {
	Id() int32

	Name() string

	SendPacket(data []byte)
}

type Party interface {
	Id() int32

	Name() string

	Players() map[int32]Player

	BroadcastPacket([]byte)

	EndGame()
}

type Game interface {
	HandleGameStarted()

	HandleGameEnded()

	HandlePlayerLeft(player Player)

	HandlePacket(sender Player, data []byte) error

	Tick()
}
