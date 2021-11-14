package main

import "github.com/veandco/go-sdl2/sdl"

type NotificationType uint8

const (
	NotificationError NotificationType = iota
)

type Notification struct {
	IsOpen  bool
	Message string
	Type    NotificationType

	MaxWidth    int32
	Padding     int32
	LineSpacing int32
}

func NewNotification() *Notification {
	return &Notification{
		IsOpen:      false,
		MaxWidth:    394,
		Padding:     8,
		LineSpacing: 3,
	}
}

func (n *Notification) Show(e NotificationEvent) {
	n.IsOpen = true
	n.Message = e.Message
	n.Type = e.Type
}

func (n *Notification) Close() {
	n.IsOpen = false
}

func (n *Notification) Tick() {

}

func (n *Notification) Render(renderer *sdl.Renderer, app *App) {
	theme := app.Theme.NotificationTheme

	lines := []string{n.Message}
	textWidth := app.Font.GetStringWidth(n.Message)
	if textWidth > (n.MaxWidth - n.Padding*2) {
		lines = app.Font.WrapString(n.Message, n.MaxWidth-n.Padding*2)
	}

	linesHeight := int32(len(lines))*app.Font.Size + n.Padding*2 + int32(len(lines)-1)*n.LineSpacing

	rect := sdl.Rect{
		X: app.WindowRect.X + (app.WindowRect.W-n.MaxWidth)/2,
		Y: app.WindowRect.Y + app.WindowRect.H - linesHeight - 10,
		W: n.MaxWidth,
		H: linesHeight,
	}

	DrawRect3D(renderer, &rect, GetColor(theme, "background_color"))

	for index, line := range lines {
		lineWidth := app.Font.GetStringWidth(line)
		lineRect := sdl.Rect{
			X: rect.X + n.Padding,
			Y: rect.Y + n.Padding + int32(index)*app.Font.Size + int32(index)*n.LineSpacing,
			W: lineWidth,
			H: app.Font.Size,
		}

		color := GetColor(theme, "error_color")

		DrawText(renderer, &app.Font, line, &lineRect, color)
	}
}