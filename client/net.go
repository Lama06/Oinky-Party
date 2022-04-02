package client

import (
	"errors"
	"fmt"
	"github.com/Lama06/Oinky-Party/protocol"
	"log"
	"net"
)

func (c *client) connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", protocol.ServerAddress, protocol.Port))
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
	defer c.disconnect()

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
	defer c.disconnect()

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
		c.disconnect()
		log.Println(errors.New("packet buffer is full"))
	}
}

func (c *client) disconnect() {
	c.disconnectOnce.Do(func() {
		err := c.conn.Close()
		if err != nil {
			log.Println(fmt.Errorf("error while closing connection to server: %w", err))
		}
	})
}
