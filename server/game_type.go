package server

import (
	"github.com/Lama06/Oinky-Party/server/connect4"
	"github.com/Lama06/Oinky-Party/server/flappyoinky"
	"github.com/Lama06/Oinky-Party/server/game"
)

var gameTypes = []gameType{
	{
		name:    "flappyoinky",
		creator: flappyoinky.Create,
	},
	{
		name:    "connect4",
		creator: connect4.Create,
	},
}

func gameTypeByName(name string) (t gameType, ok bool) {
	for _, t := range gameTypes {
		if t.name == name {
			return t, true
		}
	}

	return gameType{}, false
}

type gameType struct {
	name    string
	creator game.Creator
}
