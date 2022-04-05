package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lama06/Oinky-Party/protocol"
	"log"
	"math/rand"
	"net"
	"time"
)

func StartServer() {
	newServer().start()
}

type server struct {
	players        players
	parties        parties
	newConnections chan net.Conn
	disconnects    chan *player
}

func newServer() *server {
	return &server{
		players:        map[int32]*player{},
		parties:        map[int32]*party{},
		newConnections: make(chan net.Conn, 100),
		disconnects:    make(chan *player, 100),
	}
}

func (s *server) start() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	log.Println("server starting...")

	go s.listenForConnections()

	ticker := time.Tick(protocol.TickSpeed * time.Millisecond)
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

		err = conn.(*net.TCPConn).SetKeepAlive(true)
		if err != nil {
			log.Println(fmt.Errorf("failed to set keep alive state for connection to %s: %w", conn.RemoteAddr(), err))
		}

		s.newConnections <- conn
	}
}

func (s *server) handleNewConnection(conn net.Conn) {
	player := newPlayerForNewConnection(s, conn)

	s.players[player.id] = player

	welcome, err := json.Marshal(protocol.WelcomePacket{
		PacketName: protocol.WelcomePacketName,
		YourId:     player.id,
		YourName:   player.name,
	})
	if err != nil {
		panic(err)
	}
	player.SendPacket(welcome)

	go player.forwardMessagesFromPlayer()
	go player.forwardMessagesToPlayer()
}

func (s *server) handleDisconnect(p *player) {
	party := s.parties.byPlayer(p)
	if party != nil {
		party.removePlayer(p)
	}

	delete(s.players, p.id)
}

func (s *server) handlePacket(sender *player, data []byte) error {
	log.Printf("received packet from %s (id: %d): %s\n", sender.name, sender.id, string(data))

	packetName, err := protocol.GetPacketName(data)
	if err != nil {
		return fmt.Errorf("failed to get packet name: %w", err)
	}

	switch packetName {
	case protocol.ChangeNamePacketName:
		var changeName protocol.ChangeNamePacket
		err := json.Unmarshal(data, &changeName)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(sender)
		if currentParty != nil {
			return errors.New("cannot change name while in party")
		}

		sender.name = changeName.NewName
	case protocol.QueryPartiesPacketName:
		listParties, err := json.Marshal(protocol.ListPartiesPacket{
			PacketName: protocol.ListPartiesPacketName,
			Parties:    s.parties.toListPartiesData(),
		})
		if err != nil {
			panic(err)
		}
		sender.SendPacket(listParties)
	case protocol.CreatePartyPacketName:
		var createParty protocol.CreatePartyPacket
		err := json.Unmarshal(data, &createParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(sender)
		if currentParty != nil {
			return fmt.Errorf("player is already in a party")
		}

		party := &party{
			server:  s,
			id:      rand.Int31(),
			name:    createParty.Name,
			players: map[int32]*player{},
		}
		s.parties[party.id] = party

		party.addPlayer(sender)
	case protocol.JoinPartyPacketName:
		var joinParty protocol.JoinPartyPacket
		err := json.Unmarshal(data, &joinParty)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(sender)
		if currentParty != nil {
			return fmt.Errorf("player is already in a party")
		}

		newParty, ok := s.parties[joinParty.Id]
		if !ok {
			return fmt.Errorf("failed to find party with id: %d", joinParty.Id)
		}

		if newParty.currentGame != nil {
			return errors.New("a game is running in this party")
		}

		newParty.addPlayer(sender)
	case protocol.LeavePartyPacketName:
		currentParty := s.parties.byPlayer(sender)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}
		currentParty.removePlayer(sender)
	case protocol.StartGamePacketName:
		var startGame protocol.StartGamePacket
		err := json.Unmarshal(data, &startGame)
		if err != nil {
			return fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		currentParty := s.parties.byPlayer(sender)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}

		err = currentParty.handleStartGamePacket(startGame)
		if err != nil {
			return fmt.Errorf("failed to start game: %w", err)
		}
	case protocol.EndGamePacketName:
		currentParty := s.parties.byPlayer(sender)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}

		err := currentParty.handleEndGamePacket()
		if err != nil {
			return fmt.Errorf("failed to handle end game packet: %w", err)
		}
	default:
		currentParty := s.parties.byPlayer(sender)
		if currentParty == nil {
			return errors.New("player is not in a party")
		}
		err := currentParty.handleGamePacket(sender, data)
		if err != nil {
			return fmt.Errorf("party failed to handle the packet: %w", err)
		}
	}

	return nil
}
