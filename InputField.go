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
}

func NewInputField(rect sdl.Rect, onInput func(string)) *InputField {
	return &InputField{
		OnInputCallback:  onInput,
		Rect:             rect,
		BasePadding:      8,
		InputAreaPadding: 10,
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

func (i *InputField) Render(renderer *sdl.Renderer, x int32, y int32, font *Font, theme Subtheme) {
	DrawRect3D(renderer, &sdl.Rect{X: x, Y: y, W: i.Rect.W, H: i.Rect.H}, GetColor(theme, "background_color"))

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
	DrawRect(renderer, &cursorRect, GetColor(theme, "cursor_color"))

	if textWidth > 0 {
		DrawText(renderer, font, i.Value.String(), &textRect, GetColor(theme, "text_color"))
	}
}
