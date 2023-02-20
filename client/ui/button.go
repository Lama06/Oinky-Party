package ui

import (
	"image/color"

	"github.com/Lama06/Oinky-Party/client/rescources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	buttonPadding               = 20
	buttonScaleNotHovered       = 1
	buttonScaleHovered          = 1.2
	buttonScaleMaxChangePerTick = 0.03
)

type ButtonColorPalette struct {
	BackgroundColor      color.Color
	BackgroundHoverColor color.Color
	TextColor            color.Color
	TextHoverColor       color.Color
}

func (b ButtonColorPalette) getBackgroundHoverColor() color.Color {
	if b.BackgroundHoverColor == nil {
		return b.BackgroundColor
	}

	return b.BackgroundHoverColor
}

func (b ButtonColorPalette) getTextHoverColor() color.Color {
	if b.TextHoverColor == nil {
		return b.TextColor
	}

	return b.TextHoverColor
}

type ButtonConfig struct {
	Pos      Position
	Text     string
	Colors   *ButtonColorPalette
	Callback func()
}

type Button struct {
	Pos         Position
	Callback    func()
	text        string
	colors      *ButtonColorPalette
	scale       float64
	imgStandard *ebiten.Image
	imgHovered  *ebiten.Image
}

var _ Component = (*Button)(nil)

func NewButton(config ButtonConfig) *Button {
	button := Button{
		Pos:      config.Pos,
		Callback: config.Callback,
		text:     config.Text,
		colors:   config.Colors,
		scale:    buttonScaleNotHovered,
	}

	button.updateImages()

	return &button
}

func (b *Button) getColors() ButtonColorPalette {
	if b.colors == nil {
		return ButtonColors
	}
	return *b.colors
}

func (b *Button) getDimensions() (width int, height int) {
	bounds := text.BoundString(rescources.RobotoNormalFont, b.text)
	textWidth, textHeight := bounds.Size().X, bounds.Size().Y
	return textWidth + buttonPadding*2, textHeight + buttonPadding*2
}

func (b *Button) createImage(bgColor, textColor color.Color) *ebiten.Image {
	buttonWidth, buttonHeight := b.getDimensions()
	img := ebiten.NewImage(buttonWidth, buttonHeight)
	img.Fill(bgColor)
	textX := buttonPadding
	textY := buttonHeight - buttonPadding
	text.Draw(img, b.text, rescources.RobotoNormalFont, textX, textY, textColor)
	return img
}

func (b *Button) createImageStandard() *ebiten.Image {
	return b.createImage(b.getColors().BackgroundColor, b.getColors().TextColor)
}

func (b *Button) createImageHovered() *ebiten.Image {
	return b.createImage(b.getColors().getBackgroundHoverColor(), b.getColors().getTextHoverColor())
}

func (b *Button) updateImages() {
	b.imgStandard = b.createImageStandard()
	b.imgHovered = b.createImageHovered()
}

func (b *Button) Text() string {
	return b.text
}

func (b *Button) SetText(text string) {
	b.text = text
	b.updateImages()
}

func (b *Button) Colors() *ButtonColorPalette {
	return b.colors
}

func (b *Button) SetColors(colors *ButtonColorPalette) {
	b.colors = colors
	b.updateImages()
}

func (b *Button) Update() {
	if b.Pos == nil {
		return
	}

	buttonWidth, buttonHeight := b.getDimensions()

	if IsClicked(b.Pos, buttonWidth, buttonHeight, ebiten.MouseButtonLeft) && b.Callback != nil {
		b.Callback()
	}

	if IsHovered(b.Pos, buttonWidth, buttonHeight) {
		if b.scale < buttonScaleHovered {
			diff := buttonScaleHovered - b.scale
			if diff > buttonScaleMaxChangePerTick {
				diff = buttonScaleMaxChangePerTick
			}
			b.scale += diff
		}
	} else {
		if b.scale > buttonScaleNotHovered {
			diff := b.scale - buttonScaleNotHovered
			if diff > buttonScaleMaxChangePerTick {
				diff = buttonScaleMaxChangePerTick
			}
			b.scale -= diff
		}
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	if b.Pos == nil {
		return
	}

	buttonWidth, buttonHeight := b.getDimensions()

	img := b.imgStandard
	if IsHovered(b.Pos, buttonWidth, buttonHeight) {
		img = b.imgHovered
	}

	topLeftCornerX, topLeftCornerY := b.Pos.TopLeftCorner(buttonWidth, buttonHeight)
	var drawOptions ebiten.DrawImageOptions
	drawOptions.GeoM.Scale(b.scale, b.scale)
	drawOptions.GeoM.Translate(float64(topLeftCornerX), float64(topLeftCornerY))
	drawOptions.GeoM.Translate(-(b.scale-buttonScaleNotHovered)*0.5*float64(buttonWidth), -(b.scale-buttonScaleNotHovered)*0.5*float64(buttonHeight))
	screen.DrawImage(img, &drawOptions)
}
