package server

import (
	"fmt"
	"github.com/Lama06/Oinky-Party/protocol"
	"github.com/Lama06/Oinky-Party/server/game"
	"log"
	"net"
	"sync"
)

var randomPlayerNames = [...]string{
	"Oinky",
	"Lama",
	"Grunz Grunz",
}

type player struct {
	conn           net.Conn
	name           string
	id             int32
	send           chan []byte
	receive        chan []byte
	disconnectOnce sync.Once
	server         *server
}

var _ game.Player = (*player)(nil)

func (p *player) toData() protocol.PlayerData {
	return protocol.PlayerData{
		Name: p.name,
		Id:   p.id,
	}
}

func (p *player) forwardMessagesFromPlayer() {
	defer p.disconnect()

	for {
		msgInSizeBuffer := make([]byte, 4)
		n, err := p.conn.Read(msgInSizeBuffer)
		if err != nil || n != 4 {
			return
		}
		msgInSize := protocol.BytesToInt32([4]byte{msgInSizeBuffer[0], msgInSizeBuffer[1], msgInSizeBuffer[2], msgInSizeBuffer[3]})

		msgIn := make([]byte, 0, msgInSize)
		for len(msgIn) != int(msgInSize) {
			msgInBuffer := make([]byte, int(msgInSize)-len(msgIn))
			n, err = p.conn.Read(msgInBuffer)
			if err != nil {
				return
			}
			msgIn = append(msgIn, msgInBuffer[:n]...)
		}

		p.receive <- msgIn
	}
}

func (p *player) forwardMessagesToPlayer() {
	defer p.disconnect()

	for msgOut := range p.send {
		msgOutSize := protocol.Int32ToBytes(int32(len(msgOut)))

		_, err := p.conn.Write([]byte{msgOutSize[0], msgOutSize[1], msgOutSize[2], msgOutSize[3]})
		if err != nil {
			return
		}

		_, err = p.conn.Write(msgOut)
		if err != nil {
			return
		}
	}
}

func (p *player) disconnect() {
	p.disconnectOnce.Do(func() {
		log.Printf("disconnecting player: %s", p.name)

		err := p.conn.Close()
		if err != nil {
			log.Println(fmt.Errorf("failed to close the connection to player: %w", err))
		}

		p.server.disconnects <- p
	})
}

func (p *player) SendPacket(data []byte) {
	select {
	case p.send <- data:
		return
	default:
		log.Printf("packet buffer of player is full: %s(%d)\n", p.name, p.id)
		p.disconnect()
	}
}

func (p *player) Id() int32 {
	return p.id
}

func (p *player) Name() string {
	return p.name
}

type players map[int32]*player
