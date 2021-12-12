package main

import (
	"strings"

	"github.com/veandco/go-sdl2/ttf"
)

type Font struct {
	Data           *ttf.Font
	Size           int32
	CharacterWidth int
}

func LoadFont(path string, size int32) (result Font) {
	font, err := ttf.OpenFont(path, int(size))
	if err != nil {
		NotifyError(err.Error())
		return
	}

	// We assume that the font is going to always be monospaced
	metrics, err := font.GlyphMetrics('m')
	if err != nil {
		NotifyError(err.Error())
		return
	}

	result.Data = font
	result.Size = size
	result.CharacterWidth = metrics.Advance

	return
}

func (font *Font) GetStringWidth(text string) int32 {
	return int32(len(text) * font.CharacterWidth)
}

func (font *Font) ClipString(text string, width int32) string {
	if font.GetStringWidth(text) <= width {
		return text
	}

	maxChars := int(width / int32(font.CharacterWidth))

	var sb strings.Builder
	sb.WriteString(text[:(maxChars - 3)])
	sb.WriteString("...")

	return sb.String()
}

func (font *Font) ClipStringNoEllipsis(text string, width int32) string {
	if font.GetStringWidth(text) <= width {
		return text
	}

	maxChars := int(width / int32(font.CharacterWidth))
	return text[:maxChars]
}

func (font *Font) WrapString(text string, maxWidth int32) (result []string) {
	if font.GetStringWidth(text) <= maxWidth {
		result = []string{text}
		return
	}

	words := strings.Split(text, " ")

	var currentWidth int32 = 0

	var sb strings.Builder
	for _, word := range words {
		wordLength := font.GetStringWidth(word)
		newWidth := currentWidth + wordLength
		if newWidth < maxWidth {
			sb.WriteString(word)
			sb.WriteString(" ")
			currentWidth = newWidth + int32(font.CharacterWidth)
		} else {
			result = append(result, sb.String())
			sb.Reset()
			sb.WriteString(word)
			sb.WriteString(" ")
			currentWidth = wordLength + int32(font.CharacterWidth)
		}
	}

	result = append(result, sb.String())

	return
}

func (font *Font) Unload() {
	font.Data.Close()
}
