package server

import (
	"github.com/Lama06/Oinky-Party/server/connect4"
	"github.com/Lama06/Oinky-Party/server/flappyoinky"
	"github.com/Lama06/Oinky-Party/server/game"
	"github.com/Lama06/Oinky-Party/server/schiffe_versenken"
)

var gameTypes = []game.Type{
	flappyoinky.Type,
	connect4.Type,
	schiffe_versenken.Type,
}

func gameTypeByName(name string) (t game.Type, ok bool) {
	for _, t := range gameTypes {
		if t.Name == name {
			return t, true
		}
	}

	return game.Type{}, false
}
