package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lama06/Oinky-Party/client/flappybird"
	"github.com/Lama06/Oinky-Party/client/game"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"net"
	"sync"
)

type gameType struct {
	name        string
	displayName string
	creator     game.Creator
}

var gameTypes = []gameType{
	{
		name:        "flappybird",
		displayName: "Flappy Bird",
		creator:     flappybird.Create,
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

type client struct {
	conn           net.Conn
	send           chan []byte
	receive        chan []byte
	disconnectOnce sync.Once

	name string
	id   int32

	inParty      bool
	partyName    string
	partyId      int32
	partyPlayers []game.PartyPlayer

	currentScreen screen
	currentGame   game.Game
}

var _ game.Client = (*client)(nil)
var _ ebiten.Game = (*client)(nil)

func newClient() *client {
	return &client{
		conn:    nil,
		send:    make(chan []byte, 100),
		receive: make(chan []byte, 100),
	}
}

func (c *client) start() {
	fmt.Println("client starting...")

	log.SetFlags(log.Lshortfile | log.Ltime)

	err := c.connect()
	if err != nil {
		log.Println(fmt.Errorf("failed to connect to the server: %w", err))
		return
	}

	c.currentScreen = newMenuScreen(c)

	ebiten.SetWindowTitle("Oinky Party")
	ebiten.SetWindowResizable(true)

	err = ebiten.RunGame(c)
	if err != nil {
		log.Println(fmt.Errorf("failed to start the game: %w", err))
		return
	}
}

func (c *client) connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", protocol.Port))
	if err != nil {
		return fmt.Errorf("failed dial the server: %w", err)
	}
	err = conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		return fmt.Errorf("failed to change the keep alive state: %w", err)
	}
	c.conn = conn

	go c.forwardMessagesToServer()
	go c.forwardMessagesFromServer()

	return nil
}

func (c *client) forwardMessagesFromServer() {
	defer func() {
		err := c.disconnect()
		if err != nil {
			log.Println(fmt.Errorf("failed to close the connection to the server: %w", err))
		}
	}()

	for {
		msgInSizeBuffer := make([]byte, 4)
		n, err := c.conn.Read(msgInSizeBuffer)
		if err != nil || n != 4 {
			return
		}
		msgInSize := protocol.BytesToInt32([4]byte{msgInSizeBuffer[0], msgInSizeBuffer[1], msgInSizeBuffer[2], msgInSizeBuffer[3]})

		msgIn := make([]byte, 0, msgInSize)
		for len(msgIn) != int(msgInSize) {
			msgInBuffer := make([]byte, int(msgInSize)-len(msgIn))
			n, err = c.conn.Read(msgInBuffer)
			if err != nil {
				return
			}
			msgIn = append(msgIn, msgInBuffer[:n]...)
		}

		c.receive <- msgIn
	}
}

func (c *client) forwardMessagesToServer() {
	defer func() {
		err := c.disconnect()
		if err != nil {
			log.Println(fmt.Errorf("failed to close the connection to the server: %w", err))
		}
	}()

	for msgOut := range c.send {
		msgOutSize := protocol.Int32ToBytes(int32(len(msgOut)))

		_, err := c.conn.Write([]byte{msgOutSize[0], msgOutSize[1], msgOutSize[2], msgOutSize[3]})
		if err != nil {
			return
		}

		_, err = c.conn.Write(msgOut)
		if err != nil {
			return
		}
	}
}

func (c *client) SendPacket(packet []byte) {
	select {
	case c.send <- packet:
		return
	default:
		err := c.disconnect()
		if err != nil {
			log.Println(fmt.Errorf("failed to disconnect the player: %w", err))
		}
		log.Println(errors.New("packet buffer is full"))
	}
}

func (c *client) disconnect() error {
	var resultErr error
	c.disconnectOnce.Do(func() {
		err := c.conn.Close()
		if err != nil {
			resultErr = fmt.Errorf("error while closing connection to server: %w", err)
		}
	})
	return resultErr
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

		c.inParty = true
		c.partyName = youJoinedParty.Party.Name
		c.partyId = youJoinedParty.Party.Id
		c.partyPlayers = make([]game.PartyPlayer, len(youJoinedParty.Party.Players))
		for i, player := range youJoinedParty.Party.Players {
			c.partyPlayers[i] = game.PartyPlayer{
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

		c.inParty = false
		c.partyName = ""
		c.partyId = 0
		c.partyPlayers = nil

		c.currentScreen = newMenuScreen(c)
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
		c.partyPlayers = append(c.partyPlayers, player)
	case protocol.PlayerLeftPartyPacketName:
		var playerLeftParty protocol.PlayerLeftPartyPacket
		err := json.Unmarshal(packet, &playerLeftParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		if !c.inParty {
			return errors.New("received player left party packet but client is not in a party")
		}

		for i, player := range c.partyPlayers {
			if player.Id == playerLeftParty.Id {
				c.partyPlayers = append(c.partyPlayers[:i], c.partyPlayers[i+1:]...)
				break
			}
		}
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
		if packetHandlerComponent, ok := c.currentScreen.(packetHandlerScreen); ok {
			err := packetHandlerComponent.HandlePacket(packet)
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

func (c *client) PartyPlayers() []game.PartyPlayer {
	return c.partyPlayers
}

func (c *client) Update() error {
	for len(c.receive) != 0 {
		packet := <-c.receive
		err := c.handlePacket(packet)
		if err != nil {
			return err
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
	return outsideWidth, outsideHeight
}

func StartClient() {
	newClient().start()
}
