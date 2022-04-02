package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
	"math/rand"
)

type party struct {
	server      *server
	id          int32
	name        string
	players     []*player
	currentGame game.Game
}

var _ game.Party = (*party)(nil)

func newParty(server *server, name string) *party {
	return &party{
		server: server,
		id:     rand.Int31(),
		name:   name,
	}
}

func (p *party) toData() protocol.PartyData {
	players := make([]protocol.PlayerData, len(p.players))
	for i, player := range p.players {
		players[i] = player.toData()
	}

	return protocol.PartyData{
		Name:    p.name,
		Id:      p.id,
		Players: players,
	}
}

func (p *party) Server() game.Server {
	return p.server
}

func (p *party) BroadcastPacket(data []byte) {
	for _, player := range p.players {
		player.SendPacket(data)
	}
}

func (p *party) addPlayer(target *player) {
	playerJoinedParty, err := json.Marshal(protocol.PlayerJoinedPartyPacket{
		PacketName: protocol.PlayerJoinedPartyPacketName,
		Player:     target.toData(),
	})
	if err != nil {
		panic(err)
	}
	p.BroadcastPacket(playerJoinedParty)

	p.players = append(p.players, target)

	youJoinedParty, err := json.Marshal(protocol.YouJoinedPartyPacket{
		PacketName: protocol.YouJoinedPartyPacketName,
		Party:      p.toData(),
	})
	if err != nil {
		panic(err)
	}
	target.SendPacket(youJoinedParty)
}

func (p *party) removePlayer(target *player) {
	for i, player := range p.players {
		if player == target {
			p.players = append(p.players[:i], p.players[i+1:]...)
			break
		}
	}

	if p.currentGame != nil {
		p.currentGame.HandlePlayerLeft(target)
	}

	playerLeftParty, err := json.Marshal(protocol.PlayerLeftPartyPacket{
		PacketName: protocol.PlayerLeftPartyPacketName,
		Id:         target.id,
	})
	if err != nil {
		panic(err)
	}
	p.BroadcastPacket(playerLeftParty)

	youLeftParty, err := json.Marshal(protocol.YouLeftLeftPartyPacket{
		PacketName: protocol.YouLeftPartyPacketName,
	})
	if err != nil {
		panic(err)
	}
	target.SendPacket(youLeftParty)
}

func (p *party) startGame(t gameType) error {
	if p.currentGame != nil {
		return errors.New("a game is already running")
	}

	g := t.creator(p)
	if g == nil {
		return errors.New("cannot create the game")
	}
	g.HandleGameStarted()

	p.currentGame = g

	gameStarted, err := json.Marshal(protocol.GameStartedPacket{
		PacketName: protocol.GameStartedPacketName,
		GameType:   t.name,
	})
	if err != nil {
		panic(err)
	}
	p.BroadcastPacket(gameStarted)

	return nil
}

func (p *party) EndGame() {
	if p.currentGame == nil {
		return
	}

	p.currentGame.HandleGameEnded()
	p.currentGame = nil

	gameEnded, err := json.Marshal(protocol.GameEndedPacket{
		PacketName: protocol.GameEndedPacketName,
	})
	if err != nil {
		panic(err)
	}
	p.BroadcastPacket(gameEnded)
}

func (p *party) handlePacket(sender *player, data []byte) error {
	if p.currentGame == nil {
		return errors.New("there is no game running")
	}

	err := p.currentGame.HandlePacket(sender, data)
	if err != nil {
		return fmt.Errorf("the game failed to handle the packet: %w", err)
	}

	return nil
}

func (p *party) tick() {
	if p.currentGame != nil {
		p.currentGame.Tick()
	}
}

func (p *party) Id() int32 {
	return p.id
}

func (p *party) Name() string {
	return p.name
}

func (p *party) Players() []game.Player {
	players := make([]game.Player, len(p.players))
	for i, player := range p.players {
		players[i] = player
	}
	return players
}

type partiesManager []*party

func (p *partiesManager) byId(id int32) *party {
	for _, party := range *p {
		if party.id == id {
			return party
		}
	}

	return nil
}

func (p *partiesManager) byPlayer(target *player) *party {
	for _, party := range *p {
		for _, player := range party.players {
			if player == target {
				return party
			}
		}
	}

	return nil
}

func (p *partiesManager) toListPartiesData() []protocol.PartyData {
	parties := make([]protocol.PartyData, 0, len(*p))
	for _, party := range *p {
		if party.currentGame == nil {
			parties = append(parties, party.toData())
		}
	}
	return parties
}
