package client

import (
	"github.com/Lama06/Oinky-Party/client/flappyoinky"
	"github.com/Lama06/Oinky-Party/client/game"
)

type gameType struct {
	name        string
	displayName string
	creator     game.Creator
}

var gameTypes = []gameType{
	{
		name:        "flappyoinky",
		displayName: "Flappy Oinky",
		creator:     flappyoinky.Create,
	},
}

func gameTypeByName(name string) (gameType, bool) {
	for _, gameType := range gameTypes {
		if gameType.name == name {
			return gameType, true
		}
	}

	return gameType{}, false
}
