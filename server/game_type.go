package server

import (
	"github.com/Lama06/Oinky-Party/server/flappybird"
	"github.com/Lama06/Oinky-Party/server/game"
)

var gameTypes = []gameType{
	{
		name: "flappybird",
		creator: func(party game.Party) game.Game {
			return flappybird.Create(party)
		},
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
