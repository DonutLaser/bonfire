package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type InlineInputField struct {
	Value strings.Builder

	OnSubmit func(string)
	OnCancel func()
	IsOpen   bool

	Rect    sdl.Rect
	Padding int32
}

func NewInlineInputField() *InlineInputField {
	return &InlineInputField{
		Padding: 5,
	}
}

func (i *InlineInputField) Open(value string, submitCallback func(string), cancelCallback func()) {
	i.Value.WriteString(value)
	i.OnSubmit = submitCallback
	i.OnCancel = cancelCallback
	i.IsOpen = true
}

func (i *InlineInputField) Close() {
	i.Value.Reset()
	i.IsOpen = false
}

func (i *InlineInputField) Submit() {
	if i.OnSubmit != nil {
		i.OnSubmit(i.Value.String())
	}

	i.Close()
}

func (i *InlineInputField) Cancel() {
	if i.OnCancel != nil {
		i.OnCancel()
	}

	i.Close()
}

func (i *InlineInputField) Tick(input *Input) {
	if input.Escape {
		i.Cancel()
		return
	}

	if input.Backspace {
		if input.Ctrl {
			i.Value.Reset()
		} else {
			if i.Value.Len() > 0 {
				currentValue := i.Value.String()
				i.Value.Reset()
				i.Value.WriteString(currentValue[:len(currentValue)-1])
			}
		}

		return
	}

	if input.TypedCharacter == '\n' {
		i.Submit()
		return
	}

	if input.TypedCharacter != 0 && input.TypedCharacter != '\t' {
		i.Value.WriteByte(byte(input.TypedCharacter))
	}
}

func (i *InlineInputField) Render(renderer *sdl.Renderer, rect sdl.Rect, font *Font, theme Subtheme) {
	DrawRect3DInset(renderer, &rect, GetColor(theme, "inset_color"))

	value := i.Value.String()
	textWidth := font.GetStringWidth(value)
	textRect := sdl.Rect{
		X: rect.X + i.Padding,
		Y: rect.Y + (rect.H-font.Size)/2,
		W: textWidth,
		H: font.Size,
	}

	cursorRect := textRect
	cursorRect.X += textRect.W
	cursorRect.W = 2
	DrawRect(renderer, &cursorRect, GetColor(theme, "cursor_color"))

	if textWidth > 0 {
		DrawText(renderer, font, value, &textRect, GetColor(theme, "text_color"))
	}
}
