package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
	"log"
	"net"
	"time"
)

func StartServer() {
	newServer().start()
}

type server struct {
	players        playerManager
	parties        partiesManager
	newConnections chan net.Conn
	disconnects    chan *player
}

var _ game.Server = (*server)(nil)

func newServer() *server {
	return &server{
		newConnections: make(chan net.Conn, 100),
		disconnects:    make(chan *player, 100),
	}
}

func (s *server) start() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	log.Println("server starting...")

	go s.listenForConnections()

	ticker := time.Tick(50 * time.Millisecond)
	for {
		for len(s.newConnections) != 0 {
			s.handleNewConnection(<-s.newConnections)
		}

		for len(s.disconnects) != 0 {
			s.handleDisconnect(<-s.disconnects)
		}

		for _, player := range s.players {
			for len(player.receive) != 0 {
				err := s.handlePacket(player, <-player.receive)
				if err != nil {
					log.Println(fmt.Errorf("failed to handle packet from %s(%d): %w", player.name, player.id, err))
				}
			}
		}

		for _, party := range s.parties {
			party.tick()
		}

		<-ticker
	}
}

func (s *server) listenForConnections() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", protocol.Port))
	if err != nil {
		log.Println(fmt.Errorf("failed to start the tcp listener: %w", err))
		return
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Println(fmt.Errorf("failed to close the tcp listener: %w", err))
		}
	}()

	log.Println("listening for tcp connections...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Errorf("failed to accept a new connection: %w", err))
			continue
		}
		log.Printf("new connection from %s\n", conn.RemoteAddr())

		s.newConnections <- conn
	}
}

func (s *server) handleNewConnection(conn net.Conn) {
	player := newPlayer(conn, s)

	go player.forwardMessagesFromPlayer()
	go player.forwardMessagesToPlayer()

	welcome, err := json.Marshal(protocol.WelcomePacket{
		PacketName: protocol.WelcomePacketName,
		YourId:     player.id,
		YourName:   player.name,
	})
	if err != nil {
		panic(err)
	}
	player.SendPacket(welcome)

	s.players = append(s.players, player)
}

func (s *server) handleDisconnect(p *player) {
	party := s.parties.byPlayer(p)
	if party != nil {
		party.removePlayer(p)
	}

	s.players.remove(p)
}

func (s *server) handlePacket(player *player, data []byte) error {
	log.Printf("received packet from %s (id: %d): %s\n", player.name, player.id, string(data))

	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	switch packetName {
	case protocol.QueryPartiesPacketName:
		listParties, err := json.Marshal(protocol.ListPartiesPacket{
			PacketName: protocol.ListPartiesPacketName,
			Parties:    s.parties.toListPartiesData(),
		})
		if err != nil {
			panic(err)
		}
		player.SendPacket(listParties)
	case protocol.CreatePartyPacketName:
		var createParty protocol.CreatePartyPacket
		err := json.Unmarshal(data, &createParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(player)
		if currentParty != nil {
			return fmt.Errorf("player is already in a party")
		}

		party := newParty(s, createParty.Name)

		s.parties = append(s.parties, party)
		party.addPlayer(player)
	case protocol.JoinPartyPacketName:
		var joinParty protocol.JoinPartyPacket
		err := json.Unmarshal(data, &joinParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(player)
		if currentParty != nil {
			return fmt.Errorf("player is already in a party")
		}

		newParty := s.parties.byId(joinParty.Id)
		if newParty == nil {
			return fmt.Errorf("failed to find party with id: %d", joinParty.Id)
		}

		newParty.addPlayer(player)
	case protocol.LeavePartyPacketName:
		currentParty := s.parties.byPlayer(player)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}
		currentParty.removePlayer(player)
	case protocol.StartGamePacketName:
		var startGame protocol.StartGamePacket
		err := json.Unmarshal(data, &startGame)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		t, ok := gameTypeByName(startGame.GameType)
		if !ok {
			return fmt.Errorf("cannot find game type %s", startGame.GameType)
		}

		currentParty := s.parties.byPlayer(player)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}

		err = currentParty.startGame(t)
		if err != nil {
			return fmt.Errorf("failed to start game: %w", err)
		}
	case protocol.EndGamePacketName:
		currentParty := s.parties.byPlayer(player)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}
		currentParty.EndGame()
	default:
		currentParty := s.parties.byPlayer(player)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}
		err := currentParty.handlePacket(player, data)
		if err != nil {
			return fmt.Errorf("party failed to handle the packet: %w", err)
		}
	}

	return nil
}

func (s *server) PlayerById(id int32) game.Player {
	return s.players.byId(id)
}

func (s *server) PartyById(id int32) game.Party {
	return s.parties.byId(id)
}
