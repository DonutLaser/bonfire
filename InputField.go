package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type InputField struct {
	Value           strings.Builder
	OnInputCallback func(string)

	Rect             sdl.Rect
	BasePadding      int32
	InputAreaPadding int32

	BackgroundColor sdl.Color
	TextColor       sdl.Color
	CursorColor     sdl.Color
}

func NewInputField(rect sdl.Rect, onInput func(string)) *InputField {
	return &InputField{
		OnInputCallback:  onInput,
		Rect:             rect,
		BasePadding:      8,
		InputAreaPadding: 10,
		BackgroundColor:  sdl.Color{R: 20, G: 27, B: 39, A: 255},
		TextColor:        sdl.Color{R: 216, G: 216, B: 216, A: 255},
		CursorColor:      sdl.Color{R: 254, G: 203, B: 0, A: 255},
	}
}

func (i *InputField) Clear() {
	i.Value.Reset()
}

func (i *InputField) OnInput() {
	if i.OnInputCallback == nil {
		return
	}

	i.OnInputCallback(i.Value.String())
}

func (i *InputField) Tick(input *Input) {
	if input.Backspace {
		if input.Ctrl {
			i.Value.Reset()
		} else {
			// @TODO (!important) fix crash when string is already empty
			currentValue := i.Value.String()
			i.Value.Reset()
			i.Value.WriteString(currentValue[:len(currentValue)-1])
		}

		i.OnInput()
		return
	}

	if input.TypedCharacter != 0 {
		i.Value.WriteByte(byte(input.TypedCharacter))
		i.OnInput()
	}
}

func (i *InputField) Render(renderer *sdl.Renderer, x int32, y int32, font *Font) {
	DrawRect3D(renderer, &sdl.Rect{X: x, Y: y, W: i.Rect.W, H: i.Rect.H}, i.BackgroundColor)

	value := i.Value.String()
	textWidth := font.GetStringWidth(value)
	textRect := sdl.Rect{
		X: x + i.BasePadding + i.InputAreaPadding,
		Y: y + (i.Rect.H-font.Size)/2,
		W: textWidth,
		H: font.Size,
	}

	cursorRect := textRect
	cursorRect.X += textRect.W
	cursorRect.W = 2
	DrawRect(renderer, &cursorRect, i.CursorColor)

	if textWidth > 0 {
		DrawText(renderer, font, i.Value.String(), &textRect, i.TextColor)
	}
}
