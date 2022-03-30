package protocol

// General

const WelcomePacketName = "welcome"

type WelcomePacket struct {
	PacketName string
	YourId     int32
	YourName   string
}

const ListPartiesPacketName = "list-parties"

type ListPartiesPacket struct {
	PacketName string
	Parties    []PartyData
}

const YouJoinedPartyPacketName = "you-joined-party"

type YouJoinedPartyPacket struct {
	PacketName string
	Party      PartyData
}

const YouLeftPartyPacketName = "you-left-party"

type YouLeftLeftPartyPacket struct {
	PacketName string
}

const PlayerJoinedPartyPacketName = "player-joined-party"

type PlayerJoinedPartyPacket struct {
	PacketName string
	Player     PlayerData
}

const PlayerLeftPartyPacketName = "player-left-party"

type PlayerLeftPartyPacket struct {
	PacketName string
	Id         int32
}

const GameStartedPacketName = "game-started"

type GameStartedPacket struct {
	PacketName string
	GameType   string
}

const GameEndedPacketName = "game-ended"

type GameEndedPacket struct {
	PacketName string
}
