package connect4

const (
	BoardWidth  = 7
	BoardHeight = 6
)

type Color bool

const (
	RedColor    Color = true
	YellowColor Color = false
)

func (c Color) ToCell() Cell {
	switch c {
	case RedColor:
		return RedCell
	case YellowColor:
		return YellowCell
	default:
		panic("unreachable")
	}
}

type Cell byte

const (
	EmptyCell Cell = iota
	RedCell
	YellowCell
)

func (c Cell) ToColor() Color {
	switch c {
	case EmptyCell:
		panic("cannot convert empty cell to color")
	case RedCell:
		return RedColor
	case YellowCell:
		return YellowColor
	default:
		panic("unreachable")
	}
}

// Client zu Server

const PlacePacketName = "connect-4-player-place"

type PlacePacket struct {
	PacketName string
	X          int32
}

// Server zu Client

const PlayerPlacedPacketName = "connect-4-player-placed"

type PlayerPlacedPacket struct {
	PacketName string
	Player     Color
	X          int32
}
