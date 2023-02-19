package schiffe_versenken

import (
	"github.com/Lama06/Oinky-Party/lazy"
	shared "github.com/Lama06/Oinky-Party/schiffe_versenken"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

var borderColor = colornames.Black

func drawBorders(board *ebiten.Image) {
	width, height := board.Size()
	if width != boardWidth || height != boardHeight {
		panic("invalid board size")
	}

	pixels := make([]byte, boardWidth*boardHeight*4)
	board.ReadPixels(pixels)

	// Horizontal
	for borderIndexY := 0; borderIndexY < numberOfHorizontalBorders; borderIndexY++ {
		y := borderIndexY * (fieldSize + borderWidth)
		for x := 0; x < boardWidth; x++ {
			startIndex := (y*boardWidth + x) * 4

			pixels[startIndex] = borderColor.R
			pixels[startIndex+1] = borderColor.G
			pixels[startIndex+2] = borderColor.B
			pixels[startIndex+3] = borderColor.A
		}
	}

	// Vertikal
	for borderIndexX := 0; borderIndexX < numberOfVerticalBorders; borderIndexX++ {
		x := borderIndexX * (fieldSize + borderWidth)
		for y := 0; y < boardHeight; y++ {
			startIndex := (y*boardWidth + x) * 4

			pixels[startIndex] = borderColor.R
			pixels[startIndex+1] = borderColor.G
			pixels[startIndex+2] = borderColor.B
			pixels[startIndex+3] = borderColor.A
		}
	}

	board.WritePixels(pixels)
}

var (
	emptyFieldImg = lazy.New(func() *ebiten.Image {
		img := ebiten.NewImage(fieldSize, fieldSize)
		img.Fill(colornames.Lightblue)
		return img
	})
	shipFieldImg = lazy.New(func() *ebiten.Image {
		img := ebiten.NewImage(fieldSize, fieldSize)
		img.Fill(colornames.Green)
		return img
	})
)

func drawShips(board *ebiten.Image, ships [shared.BoardWidth][shared.BoardHeight]bool) {
	width, height := board.Size()
	if width != boardWidth || height != boardHeight {
		panic("invalid board size")
	}

	for x := 0; x < shared.BoardWidth; x++ {
		for y := 0; y < shared.BoardHeight; y++ {
			var fieldImg *ebiten.Image
			switch ships[x][y] {
			case false:
				fieldImg = emptyFieldImg()
			case true:
				fieldImg = shipFieldImg()
			}

			var fieldImgDrawOptions ebiten.DrawImageOptions
			fieldImgDrawOptions.GeoM.Translate(float64(x*fieldSize+(x+1)*borderWidth), float64(y*fieldSize+(y+1)*borderWidth))
			board.DrawImage(fieldImg, &fieldImgDrawOptions)
		}
	}
}

var (
	markerColor = colornames.Black
	markerImg   = lazy.New(func() *ebiten.Image {
		img := ebiten.NewImage(fieldSize, fieldSize)

		pixels := make([]byte, fieldSize*fieldSize*4)
		img.ReadPixels(pixels)

		for y := 0; y < fieldSize; y++ {
			x := y

			startIndex := (y*fieldSize + x) * 4
			pixels[startIndex] = markerColor.R
			pixels[startIndex+1] = markerColor.G
			pixels[startIndex+2] = markerColor.B
			pixels[startIndex+3] = markerColor.A

			x = fieldSize - 1 - y

			startIndex = (y*fieldSize + x) * 4
			pixels[startIndex] = markerColor.R
			pixels[startIndex+1] = markerColor.G
			pixels[startIndex+2] = markerColor.B
			pixels[startIndex+3] = markerColor.A
		}

		img.WritePixels(pixels)

		return img
	})
)

func drawMarkers(board *ebiten.Image, markers [shared.BoardWidth][shared.BoardHeight]bool) {
	width, height := board.Size()
	if width != boardWidth || height != boardHeight {
		panic("invalid board size")
	}

	for x := 0; x < shared.BoardWidth; x++ {
		for y := 0; y < shared.BoardHeight; y++ {
			if !markers[x][y] {
				continue
			}

			var markerImgDrawOptions ebiten.DrawImageOptions
			markerImgDrawOptions.GeoM.Translate(float64(x*fieldSize+(x+1)*borderWidth), float64(y*fieldSize+(y+1)*borderWidth))
			board.DrawImage(markerImg(), &markerImgDrawOptions)
		}
	}
}

func getFieldCoordinates(mouseX, mouseY int) (int, int, bool) {
	fieldX := mouseX / (borderWidth + fieldSize)
	fieldY := mouseY / (borderWidth + fieldSize)

	if fieldX < 0 || fieldX >= shared.BoardWidth || fieldY < 0 || fieldY >= shared.BoardHeight {
		return 0, 0, false
	}

	return fieldX, fieldY, true
}
