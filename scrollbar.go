package main

import "github.com/veandco/go-sdl2/sdl"

type Scrollbar struct {
	Rect sdl.Rect
}

func NewScrollbar(rect sdl.Rect) *Scrollbar {
	return &Scrollbar{Rect: rect}
}

func (s *Scrollbar) Resize(rect sdl.Rect) {
	s.Rect = rect
}

func (s *Scrollbar) Render(renderer *sdl.Renderer, progress int32, max int32, app *App) {
	theme := app.Theme.PreviewTheme

	DrawRect3DInset(renderer, &s.Rect, GetColor(theme, "inset_color"))

	handleWidth := s.Rect.W / max
	handleRect := sdl.Rect{
		X: s.Rect.X + 1 + progress*handleWidth,
		Y: s.Rect.Y + 1,
		W: s.Rect.W / max,
		H: s.Rect.H - 2,
	}

	DrawRect3D(renderer, &handleRect, GetColor(theme, "background_color"))
}
