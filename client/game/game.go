package game

import "github.com/hajimehoshi/ebiten/v2"

type Type struct {
	Creator     Creator
	Name        string
	DisplayName string
}

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

	PartyPlayers() map[int32]PartyPlayer

	SendPacket(packet []byte)
}

type Game interface {
	HandleGameStarted()

	HandleGameEnded()

	HandlePacket(packet []byte) error

	Draw(screen *ebiten.Image)

	Update()

	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}
