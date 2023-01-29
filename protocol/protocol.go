package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	Port      = 3333
	TickSpeed = 50
)

type NamedPacket struct {
	PacketName string
}

func GetPacketName(data []byte) (string, error) {
	var named NamedPacket
	err := json.Unmarshal(data, &named)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return named.PacketName, nil
}

type PlayerData struct {
	Name string
	Id   int32
}

type PartyData struct {
	Name    string
	Id      int32
	Players []PlayerData
}

func Int32ToBytes(n int32) [4]byte {
	return [4]byte{
		byte((n >> 24) & 0xff),
		byte((n >> 16) & 0xff),
		byte((n >> 8) & 0xff),
		byte(n & 0xff),
	}
}

func BytesToInt32(bytes [4]byte) int32 {
	return int32(bytes[0])<<24 + int32(bytes[1])<<16 + int32(bytes[2])<<8 + int32(bytes[3])
}
