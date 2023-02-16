package client

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Lama06/Oinky-Party/client/game"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
)

func StartClient() {
	newClient().start()
}

type client struct {
	conn           net.Conn
	send           chan []byte
	receive        chan []byte
	disconnected   chan struct{}
	disconnectOnce sync.Once

	name string
	id   int32

	inParty      bool
	partyName    string
	partyId      int32
	partyPlayers map[int32]game.PartyPlayer

	currentScreen screen
	currentGame   game.Game
}

var _ game.Client = (*client)(nil)
var _ ebiten.Game = (*client)(nil)

func newClient() *client {
	return &client{
		send:         make(chan []byte, 100),
		receive:      make(chan []byte, 100),
		disconnected: make(chan struct{}, 1),
		partyPlayers: map[int32]game.PartyPlayer{},
	}
}

func (c *client) start() {
	serverAddress := flag.String("address", "localhost", "Server Address")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ltime)
	log.Println("client starting...")

	err := c.connect(*serverAddress)
	if err != nil {
		log.Println(fmt.Errorf("failed to connect to the server: %w", err))
		return
	}

	ebiten.SetWindowTitle("Oinky Party")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	c.currentScreen = newTitleScreen(c)

	err = ebiten.RunGame(c)
	if err != nil {
		log.Println(fmt.Errorf("failed to start the game: %w", err))
		return
	}
}

func (c *client) handlePacket(packet []byte) error {
	log.Println("received packet:", string(packet))

	packetName, err := protocol.GetPacketName(packet)
	if err != nil {
		return fmt.Errorf("failed to obtain the name of the packet: %w", err)
	}

	switch packetName {
	case protocol.WelcomePacketName:
		var welcome protocol.WelcomePacket
		err := json.Unmarshal(packet, &welcome)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		c.id = welcome.YourId
		c.name = welcome.YourName
	case protocol.YouJoinedPartyPacketName:
		var youJoinedParty protocol.YouJoinedPartyPacket
		err := json.Unmarshal(packet, &youJoinedParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if c.inParty {
			return errors.New("already in a party")
		}

		c.inParty = true
		c.partyName = youJoinedParty.Party.Name
		c.partyId = youJoinedParty.Party.Id
		c.partyPlayers = make(map[int32]game.PartyPlayer, len(youJoinedParty.Party.Players))
		for _, player := range youJoinedParty.Party.Players {
			c.partyPlayers[player.Id] = game.PartyPlayer{
				Name: player.Name,
				Id:   player.Id,
			}
		}

		c.currentScreen = newPartyScreen(c)
	case protocol.YouLeftPartyPacketName:
		var youLeftParty protocol.YouLeftLeftPartyPacket
		err := json.Unmarshal(packet, &youLeftParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("not in a party")
		}

		c.inParty = false
		c.partyName = ""
		c.partyId = 0
		c.partyPlayers = nil

		c.currentScreen = newTitleScreen(c)
	case protocol.PlayerJoinedPartyPacketName:
		var playerJoinedParty protocol.PlayerJoinedPartyPacket
		err := json.Unmarshal(packet, &playerJoinedParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("received player joined party packet but client is not in a party")
		}

		player := game.PartyPlayer{
			Name: playerJoinedParty.Player.Name,
			Id:   playerJoinedParty.Player.Id,
		}
		c.partyPlayers[player.Id] = player
	case protocol.PlayerLeftPartyPacketName:
		var playerLeftParty protocol.PlayerLeftPartyPacket
		err := json.Unmarshal(packet, &playerLeftParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("received player left party packet but client is not in a party")
		}

		delete(c.partyPlayers, playerLeftParty.Id)
	case protocol.GameStartedPacketName:
		var gameStarted protocol.GameStartedPacket
		err := json.Unmarshal(packet, &gameStarted)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("received game started packet but client is not in a party")
		}

		if c.currentGame != nil {
			return errors.New("received game started packet but there is a game already running")
		}

		gameType, ok := gameTypeByName(gameStarted.GameType)
		if !ok {
			return fmt.Errorf("unknown game type: %s", gameStarted.GameType)
		}

		newGame := gameType.creator(c)
		newGame.HandleGameStarted()
		c.currentGame = newGame
		c.currentScreen = newGameScreen(c)
	case protocol.GameEndedPacketName:
		var gameEnded protocol.GameEndedPacket
		err := json.Unmarshal(packet, &gameEnded)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("received game ended packet but client is not in a party")
		}

		if c.currentGame == nil {
			return errors.New("received game ended packet but there is no game running")
		}

		c.currentGame.HandleGameEnded()
		c.currentGame = nil
		c.currentScreen = newPartyScreen(c)
	default:
		if packetHandler, ok := c.currentScreen.(packetHandlerScreen); ok {
			err := packetHandler.HandlePacket(packet)
			if err != nil {
				return fmt.Errorf("screen failed to handle error: %w", err)
			}
		}
	}

	return nil
}

func (c *client) Name() string {
	return c.name
}

func (c *client) Id() int32 {
	return c.id
}

func (c *client) PartyName() string {
	return c.partyName
}

func (c *client) PartyId() int32 {
	return c.partyId
}

func (c *client) PartyPlayers() map[int32]game.PartyPlayer {
	partyPlayers := make(map[int32]game.PartyPlayer, len(c.partyPlayers))
	for id, player := range c.partyPlayers {
		partyPlayers[id] = player
	}
	return partyPlayers
}

func (c *client) Update() error {
	if len(c.disconnected) == 1 {
		return errors.New("disconnected from the server")
	}

	for len(c.receive) != 0 {
		packet := <-c.receive
		err := c.handlePacket(packet)
		if err != nil {
			log.Println(fmt.Errorf("failed to handle packet from server: %w", err))
		}
	}

	if c.currentScreen != nil {
		c.currentScreen.Update()
	}

	return nil
}

func (c *client) Draw(screen *ebiten.Image) {
	if c.currentScreen != nil {
		c.currentScreen.Draw(screen)
	}
}

func (c *client) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if c.currentGame != nil {
		return c.currentGame.Layout(outsideWidth, outsideHeight)
	}
	return outsideWidth, outsideHeight
}
