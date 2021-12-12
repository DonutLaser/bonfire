package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type Breadcrumbs struct {
	Path            []string
	AvailableDrives []string

	Rect       sdl.Rect
	ShowDrives bool
}

func NewBreadcrumbs(rect sdl.Rect, availableDrives []string) *Breadcrumbs {
	return &Breadcrumbs{
		Path:            []string{},
		AvailableDrives: availableDrives,
		Rect:            rect,
	}
}

func (b *Breadcrumbs) Clear() {
	b.Path = []string{}
}

func (b *Breadcrumbs) Push(path string) {
	b.Path = append(b.Path, path)
}

func (b *Breadcrumbs) Pop() string {
	if len(b.Path) <= 1 {
		return ""
	}

	result := b.Path[len(b.Path)-1]
	b.Path = b.Path[:len(b.Path)-1]
	return result
}

func (b *Breadcrumbs) Set(fullPath string) {
	b.Path = strings.Split(fullPath, "/")
}

func (b *Breadcrumbs) ShowAvailableDrives(show bool) {
	b.ShowDrives = show
}

func (b *Breadcrumbs) Resize(rect sdl.Rect) {
	b.Rect = rect
}

func (b *Breadcrumbs) Render(renderer *sdl.Renderer, font *Font, theme Subtheme) {
	DrawRect3D(renderer, &b.Rect, GetColor(theme, "background_color"))

	if b.ShowDrives {
		staticText := "Available drives: "
		drives := strings.Join(b.AvailableDrives, ", ")

		staticTextWidth := font.GetStringWidth(staticText)
		drivesWidth := font.GetStringWidth(drives)

		staticRect := sdl.Rect{
			X: b.Rect.X + (b.Rect.W-staticTextWidth-drivesWidth)/2,
			Y: b.Rect.Y + (b.Rect.H-font.Size)/2,
			W: staticTextWidth,
			H: font.Size,
		}
		drivesRect := sdl.Rect{
			X: staticRect.X + staticRect.W,
			Y: b.Rect.Y + (b.Rect.H-font.Size)/2,
			W: drivesWidth,
			H: font.Size,
		}

		DrawText(renderer, font, staticText, &staticRect, GetColor(theme, "separator_color"))
		DrawText(renderer, font, drives, &drivesRect, GetColor(theme, "path_color"))
	} else if len(b.Path) > 0 {
		symbol := GetString(theme, "separator_symbol")

		fullPath := strings.Join(b.Path, symbol)
		pathWidth := font.GetStringWidth(fullPath)

		cursorX := b.Rect.X + (b.Rect.W-pathWidth)/2
		cursorY := b.Rect.Y + (b.Rect.H-font.Size)/2

		joinWidth := font.GetStringWidth(symbol)

		for index, path := range b.Path {
			width := font.GetStringWidth(path)

			DrawText(renderer, font, path, &sdl.Rect{X: cursorX, Y: cursorY, W: width, H: font.Size}, GetColor(theme, "path_color"))
			if index < len(b.Path)-1 {
				DrawText(renderer, font, symbol, &sdl.Rect{X: cursorX + width, Y: cursorY, W: joinWidth, H: font.Size}, GetColor(theme, "separator_color"))
				cursorX += joinWidth
			}

			cursorX += width
		}
	}
}
