package schiffe_versenken

import "sort"

const (
	Name        = "schiffe_versenken"
	BoardWidth  = 10
	BoardHeight = 10
)

var NumberOfShips = map[int]int{
	1: 1,
	2: 2,
}

type Position struct {
	X, Y int
}

func (p Position) Valid() bool {
	return p.X >= 0 && p.X < BoardWidth && p.Y >= 0 && p.Y < BoardHeight
}

func (p Position) Neighbours() map[Position]struct{} {
	result := make(map[Position]struct{})
	for _, xOffset := range []int{-1, 0, 1} {
		for _, yOffset := range []int{-1, 0, 1} {
			neighbour := Position{p.X + xOffset, p.Y + yOffset}
			if neighbour.Valid() && neighbour != p {
				result[neighbour] = struct{}{}
			}
		}
	}
	return result
}

type Ship []Position

func areIntsEqual(ints []int) bool {
	if len(ints) == 0 {
		return true
	}

	commonValue := ints[0]
	for _, value := range ints {
		if value != commonValue {
			return false
		}
	}

	return true
}

func areIntsContiguous(ints []int) bool {
	if len(ints) == 0 {
		return true
	}

	sort.Ints(ints)

	last := ints[0]
	for i := 1; i < len(ints); i++ {
		current := ints[i]
		if last+1 != current {
			return false
		}
		last = current
	}

	return true
}

func (s Ship) Valid() bool {
	for _, pos := range s {
		if !pos.Valid() {
			return false
		}
	}

	xPositions := make([]int, len(s))
	for i, pos := range s {
		xPositions[i] = pos.X
	}

	yPositions := make([]int, len(s))
	for i, pos := range s {
		yPositions[i] = pos.Y
	}

	return (areIntsEqual(xPositions) && areIntsContiguous(yPositions)) ||
		(areIntsEqual(yPositions) && areIntsContiguous(xPositions))
}

func (s Ship) blockedFields() map[Position]struct{} {
	result := make(map[Position]struct{})
	for _, pos := range s {
		result[pos] = struct{}{}
		for neighbour := range pos.Neighbours() {
			result[neighbour] = struct{}{}
		}
	}
	return result
}

type Ships []Ship

func (s Ships) countShipsWithLength(length int) (result int) {
	for _, ship := range s {
		if len(ship) == length {
			result++
		}
	}
	return
}

func (s Ships) Valid() bool {
	totalShipCountShould := 0
	for _, count := range NumberOfShips {
		totalShipCountShould += count
	}
	if totalShipCountShould != len(s) {
		return false
	}

	for shipLength, count := range NumberOfShips {
		if s.countShipsWithLength(shipLength) != count {
			return false
		}
	}

	for shipIndex, ship := range s {
		for otherShipIndex, otherShip := range s {
			if shipIndex == otherShipIndex {
				continue
			}

			for shipBlockedField := range ship.blockedFields() {
				for _, otherShipField := range otherShip {
					if shipBlockedField == otherShipField {
						return false
					}
				}
			}
		}
	}

	return true
}

const packetNamePrefix = "schiffe-versenken-"

// Client zu Server

const SetupShipsPacketName = packetNamePrefix + "setup-ships"

type SetupShipsPacket struct {
	PacketName string
	Ships      Ships
}

const FirePacketName = packetNamePrefix + "fire"

type FirePacket struct {
	PacketName string
	Position   Position
}

// Server zu Client

const GameStartedPacketName = packetNamePrefix + "game-started"

type GameStartedPacket struct {
	PacketName string
}

const FireResultPacketName = packetNamePrefix + "fire-result"

type FireResult byte

type FireResultPacket struct {
	PacketName string
	Position   Position
	Hit        bool
}

const OpponentFiredPacketName = packetNamePrefix + "ship-destroyed"

type OpponentFiredPacket struct {
	PacketName string
	Position   Position
}
