package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Info struct {
	Name     string
	Size     string
	Created  string
	Modified string
}

type InfoView struct {
	IsOpen bool
	Info

	MaxWidth     int32
	Padding      int32
	ItemPadding  int32
	HeaderHeight int32
	ItemHeight   int32
}

func NewInfoView() *InfoView {
	return &InfoView{
		IsOpen:       false,
		MaxWidth:     300,
		Padding:      8,
		ItemPadding:  5,
		HeaderHeight: 28,
		ItemHeight:   24,
	}
}

func (i *InfoView) Show(info Info) {
	i.IsOpen = true
	i.Info = info
}

func (i *InfoView) Close() {
	i.IsOpen = false
}

func (i *InfoView) Tick(input *Input) {
	if input.Escape {
		i.Close()
	}
}

func (i *InfoView) Render(renderer *sdl.Renderer, app *App) {
	theme := app.Theme.InfoViewTheme

	headerRect := sdl.Rect{
		X: app.WindowRect.X + app.WindowRect.W - 10 - i.MaxWidth,
		Y: app.WindowRect.Y + app.WindowRect.H - 10 - i.HeaderHeight - i.ItemHeight*3 - i.Padding*2,
		W: i.MaxWidth,
		H: i.HeaderHeight,
	}
	DrawRect3D(renderer, &headerRect, GetColor(theme, "background_color"))

	clippedName := app.Font.ClipString(i.Name, i.MaxWidth-i.Padding*2)

	nameWidth := app.Font.GetStringWidth(clippedName)
	nameRect := sdl.Rect{
		X: headerRect.X + 10,
		Y: headerRect.Y + (headerRect.H-app.Font.Size)/2,
		W: nameWidth,
		H: app.Font.Size,
	}
	DrawText(renderer, &app.Font, clippedName, &nameRect, GetColor(theme, "header_color"))

	baseRect := sdl.Rect{
		X: app.WindowRect.X + app.WindowRect.W - 10 - i.MaxWidth,
		Y: headerRect.Y + headerRect.H,
		W: i.MaxWidth,
		H: i.ItemHeight*3 + i.Padding*2,
	}
	insetRect := sdl.Rect{
		X: baseRect.X + i.Padding,
		Y: baseRect.Y + i.Padding,
		W: baseRect.W - i.Padding*2,
		H: baseRect.H - i.Padding*2,
	}
	DrawRect3D(renderer, &baseRect, GetColor(theme, "background_color"))
	DrawRect3DInset(renderer, &insetRect, GetColor(theme, "inset_color"))

	sizePropWidth := app.Font.GetStringWidth("Size")
	sizeWidth := app.Font.GetStringWidth(i.Info.Size)
	sizePropRect := sdl.Rect{
		X: insetRect.X + i.ItemPadding,
		Y: insetRect.Y + (i.ItemHeight-app.Font.Size)/2,
		W: sizePropWidth,
		H: app.Font.Size,
	}
	sizeValueRect := sdl.Rect{
		X: insetRect.X + insetRect.W - i.ItemPadding - sizeWidth,
		Y: insetRect.Y + (i.ItemHeight-app.Font.Size)/2,
		W: sizeWidth,
		H: app.Font.Size,
	}
	DrawText(renderer, &app.Font, "Size", &sizePropRect, GetColor(theme, "info_color"))
	DrawText(renderer, &app.Font, i.Info.Size, &sizeValueRect, GetColor(theme, "info_color"))

	createPropoWidth := app.Font.GetStringWidth("Created")
	createWidth := app.Font.GetStringWidth(i.Info.Created)
	createPropRect := sdl.Rect{
		X: insetRect.X + i.ItemPadding,
		Y: insetRect.Y + i.ItemHeight*2 + (i.ItemHeight-app.Font.Size)/2,
		W: createPropoWidth,
		H: app.Font.Size,
	}
	createValueRect := sdl.Rect{
		X: insetRect.X + insetRect.W - i.ItemPadding - createWidth,
		Y: insetRect.Y + i.ItemHeight*2 + (i.ItemHeight-app.Font.Size)/2,
		W: createWidth,
		H: app.Font.Size,
	}
	DrawText(renderer, &app.Font, "Created", &createPropRect, GetColor(theme, "info_color"))
	DrawText(renderer, &app.Font, i.Info.Modified, &createValueRect, GetColor(theme, "info_color"))

	modPropWidth := app.Font.GetStringWidth("Modified")
	modWidth := app.Font.GetStringWidth(i.Info.Modified)
	modPropRect := sdl.Rect{
		X: insetRect.X + i.ItemPadding,
		Y: insetRect.Y + i.ItemHeight + (i.ItemHeight-app.Font.Size)/2,
		W: modPropWidth,
		H: app.Font.Size,
	}
	modValueRect := sdl.Rect{
		X: insetRect.X + insetRect.W - i.ItemPadding - modWidth,
		Y: insetRect.Y + i.ItemHeight + (i.ItemHeight-app.Font.Size)/2,
		W: modWidth,
		H: app.Font.Size,
	}
	DrawText(renderer, &app.Font, "Modified", &modPropRect, GetColor(theme, "info_color"))
	DrawText(renderer, &app.Font, i.Info.Modified, &modValueRect, GetColor(theme, "info_color"))
}
