package game

import "github.com/hajimehoshi/ebiten/v2"

type Creator func(client Client) Game

type PartyPlayer struct {
	Name string
	Id   int32
}

type Client interface {
	Name() string

	Id() int32

	PartyName() string

	PartyId() int32

	PartyPlayers() []PartyPlayer

	SendPacket(packet []byte)
}

type Game interface {
	HandleGameStarted()

	HandleGameEnded()

	HandlePlayerLeft()

	HandlePacket(packet []byte) error

	Draw(screen *ebiten.Image)

	Update()
}