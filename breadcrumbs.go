package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type Breadcrumbs struct {
	Path []string

	Rect sdl.Rect

	BackgroundColor sdl.Color
	PathColor       sdl.Color
	JoinColor       sdl.Color
}

func NewBreadcrumbs(rect sdl.Rect) *Breadcrumbs {
	return &Breadcrumbs{
		Path:            []string{},
		Rect:            rect,
		BackgroundColor: sdl.Color{R: 20, G: 27, B: 39, A: 255},
		PathColor:       sdl.Color{R: 254, G: 203, B: 0, A: 255},
		JoinColor:       sdl.Color{R: 216, G: 216, B: 216, A: 255},
	}
}

func (b *Breadcrumbs) Clear() {
	b.Path = []string{}
}

func (b *Breadcrumbs) Push(path string) {
	b.Path = append(b.Path, path)
}

func (b *Breadcrumbs) Pop() string {
	result := b.Path[len(b.Path)-1]
	b.Path = b.Path[:len(b.Path)-1]
	return result
}

func (b *Breadcrumbs) Resize(rect sdl.Rect) {
	b.Rect = rect
}

func (b *Breadcrumbs) Render(renderer *sdl.Renderer, font *Font) {
	DrawRect3D(renderer, &b.Rect, b.BackgroundColor)

	if len(b.Path) > 0 {
		fullPath := strings.Join(b.Path, " > ")
		pathWidth := font.GetStringWidth(fullPath)

		cursorX := b.Rect.X + (b.Rect.W-pathWidth)/2
		cursorY := b.Rect.Y + (b.Rect.H-font.Size)/2

		joinWidth := font.GetStringWidth(" > ")

		for index, path := range b.Path {
			width := font.GetStringWidth(path)

			DrawText(renderer, font, path, &sdl.Rect{X: cursorX, Y: cursorY, W: width, H: font.Size}, b.PathColor)
			if index < len(b.Path)-1 {
				DrawText(renderer, font, " > ", &sdl.Rect{X: cursorX + width, Y: cursorY, W: joinWidth, H: font.Size}, b.JoinColor)
				cursorX += joinWidth
			}

			cursorX += width
		}
	}
}
