package schiffe_versenken

import (
	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type setupBoard struct {
	game  *impl
	ships [shared.BoardWidth][shared.BoardHeight]bool
}

func newEmptySetupBoard(game *impl) *setupBoard {
	return &setupBoard{
		game: game,
	}
}

type parseShipsDirection byte

const (
	horizontalParseDirection parseShipsDirection = iota
	verticalParseDirection
)

func (p parseShipsDirection) nextPosition(position shared.Position) (shared.Position, bool) {
	next := position
	switch p {
	case horizontalParseDirection:
		next.X++
	case verticalParseDirection:
		next.Y++
	}
	return next, next.Valid()
}

func (s *setupBoard) parseShipsInLine(
	startPos shared.Position,
	direction parseShipsDirection,
) (ships []shared.Ship) {
	var currentShip shared.Ship
	currentPosition := startPos
	for {
		if s.ships[currentPosition.X][currentPosition.Y] {
			currentShip = append(currentShip, currentPosition)
		} else {
			if len(currentShip) > 1 {
				ships = append(ships, currentShip)
			}
			currentShip = nil
		}

		nextPosition, ok := direction.nextPosition(currentPosition)
		if !ok {
			break
		}
		currentPosition = nextPosition
	}
	if len(currentShip) > 1 {
		ships = append(ships, currentShip)
	}
	return
}

func (s *setupBoard) parseHorizontalShips() (ships []shared.Ship) {
	for y := 0; y < shared.BoardHeight; y++ {
		ships = append(ships, s.parseShipsInLine(shared.Position{X: 0, Y: y}, horizontalParseDirection)...)
	}
	return
}

func (s *setupBoard) parseVerticalShips() (ships []shared.Ship) {
	for x := 0; x < shared.BoardWidth; x++ {
		ships = append(ships, s.parseShipsInLine(shared.Position{X: x, Y: 0}, verticalParseDirection)...)
	}
	return
}

func (s *setupBoard) parseSingleFieldShips() (ships []shared.Ship) {
	for x := 0; x < shared.BoardWidth; x++ {
	yLoop:
		for y := 0; y < shared.BoardHeight; y++ {
			if !s.ships[x][y] {
				continue
			}

			position := shared.Position{X: x, Y: y}
			for neighbour := range position.Neighbours() {
				if s.ships[neighbour.X][neighbour.Y] {
					continue yLoop
				}
			}

			ships = append(ships, shared.Ship{position})
		}
	}
	return
}

func (s *setupBoard) parseShips() (ships shared.Ships) {
	for _, ship := range s.parseSingleFieldShips() {
		ships = append(ships, ship)
	}
	for _, ship := range s.parseHorizontalShips() {
		ships = append(ships, ship)
	}
	for _, ship := range s.parseVerticalShips() {
		ships = append(ships, ship)
	}
	return
}

func (s *setupBoard) draw(screen *ebiten.Image) {
	board := ebiten.NewImage(boardWidth, boardHeight)
	drawBorders(board)
	drawShips(board, s.ships)
	screen.DrawImage(board, nil)
}

func (s *setupBoard) update() {
	fieldX, fieldY, ok := getFieldCoordinates(ebiten.CursorPosition())
	if !ok {
		return
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		s.ships[fieldX][fieldY] = !s.ships[fieldX][fieldY]
	}
}
