package game

type Creator func(party Party) Game

type Server interface {
	PlayerById(id int32) Player

	PartyById(id int32) Party
}

type Player interface {
	Id() int32

	Name() string

	SendPacket(data []byte)
}

type Party interface {
	Server() Server

	Id() int32

	Name() string

	Players() []Player

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
