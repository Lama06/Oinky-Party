package client

import (
	"github.com/Lama06/Oinky-Party/client/connect4"
	"github.com/Lama06/Oinky-Party/client/flappyoinky"
	"github.com/Lama06/Oinky-Party/client/game"
)

var gameTypes = []game.Type{flappyoinky.Type, connect4.Type}

func gameTypeByName(name string) (game.Type, bool) {
	for _, gameType := range gameTypes {
		if gameType.Name == name {
			return gameType, true
		}
	}

	return game.Type{}, false
}
