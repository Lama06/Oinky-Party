package client

import (
	"github.com/Lama06/Oinky-Party/client/connect4"
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
	{
		name:        "connect4",
		displayName: "Connect 4",
		creator:     connect4.Create,
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
