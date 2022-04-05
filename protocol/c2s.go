package protocol

const ChangeNamePacketName = "change-name"

type ChangeNamePacket struct {
	PacketName string
	NewName    string
}

const CreatePartyPacketName = "create-party"

type CreatePartyPacket struct {
	PacketName string
	Name       string
}

const QueryPartiesPacketName = "query-parties"

type QueryPartiesPacket struct {
	PacketName string
}

const JoinPartyPacketName = "join-party"

type JoinPartyPacket struct {
	PacketName string
	Id         int32
}

const LeavePartyPacketName = "leave-party"

type LeavePartyPacket struct {
	PacketName string
}

const StartGamePacketName = "start-game"

type StartGamePacket struct {
	PacketName string
	GameType   string
}

const EndGamePacketName = "end-game"

type EndGamePacket struct {
	PacketName string
}
