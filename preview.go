package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type PreviewMode string

const (
	PreviewModeImage       PreviewMode = "image"
	PreviewModeText        PreviewMode = "text"
	PreviewModeUnsupported PreviewMode = "unsupported"
)

type Preview struct {
	IsOpen      bool
	PreviewMode PreviewMode

	Name  string
	Image *Image
	Text  string

	Padding      int32
	HeaderHeight int32
}

func NewPreview() *Preview {
	return &Preview{
		PreviewMode:  PreviewModeUnsupported,
		Padding:      8,
		HeaderHeight: 28,
	}
}

func (p *Preview) ShowImage(name string, image *Image) {
	p.Name = name
	p.Image = image

	p.PreviewMode = PreviewModeImage
	p.IsOpen = true
}

func (p *Preview) ShowText(name string, text string) {
	p.Name = name
	p.Text = text

	p.PreviewMode = PreviewModeText
	p.IsOpen = true
}

func (p *Preview) ShowPreviewUnsupported(name string) {
	p.Name = name

	p.PreviewMode = PreviewModeUnsupported
	p.IsOpen = true
}

func (p *Preview) Close() {
	p.IsOpen = false
}

func (p *Preview) Tick(input *Input) {
	if input.Escape {
		p.Close()
		return
	}
}

func (p *Preview) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, app *App) {
	theme := app.Theme.PreviewTheme

	headerRect := sdl.Rect{
		X: parentRect.X + parentRect.W/2,
		Y: parentRect.Y + app.Breadcrumbs[app.ActiveView].Rect.H + p.Padding,
		W: parentRect.W/2 - p.Padding,
		H: p.HeaderHeight,
	}
	DrawRect3D(renderer, &headerRect, GetColor(theme, "background_color"))

	clippedName := app.Font.ClipString("Preview: "+p.Name, headerRect.W-p.Padding)

	nameWidth := app.Font.GetStringWidth(clippedName)
	nameRect := sdl.Rect{
		X: headerRect.X + 10,
		Y: headerRect.Y + (headerRect.H-app.Font.Size)/2,
		W: nameWidth,
		H: app.Font.Size,
	}
	DrawText(renderer, &app.Font, clippedName, &nameRect, GetColor(theme, "header_color"))

	baseRect := sdl.Rect{
		X: headerRect.X,
		Y: headerRect.Y + p.HeaderHeight,
		W: headerRect.W,
		H: parentRect.H - app.Breadcrumbs[app.ActiveView].Rect.H - p.HeaderHeight - p.Padding*2,
	}
	insetRect := sdl.Rect{
		X: baseRect.X + p.Padding,
		Y: baseRect.Y + p.Padding,
		W: baseRect.W - p.Padding*2,
		H: baseRect.H - p.Padding*2,
	}
	DrawRect3D(renderer, &baseRect, GetColor(theme, "background_color"))
	DrawRect3DInset(renderer, &insetRect, GetColor(theme, "inset_color"))

	if p.PreviewMode == PreviewModeImage {
		ratio := float32(p.Image.Height) / float32(p.Image.Width)
		newWidth := insetRect.W - p.Padding*2
		newHeight := int32(float32(newWidth) * ratio)

		imageRect := sdl.Rect{
			X: insetRect.X + p.Padding,
			Y: insetRect.Y + (insetRect.H-newHeight)/2,
			W: newWidth,
			H: newHeight,
		}

		if p.Image.Height > p.Image.Width {
			ratio = float32(p.Image.Width) / float32(p.Image.Height)
			newHeight = insetRect.H - p.Padding*2
			newWidth = int32(float32(newHeight) * ratio)

			imageRect = sdl.Rect{
				X: insetRect.X + (insetRect.W-newWidth)/2,
				Y: insetRect.Y + p.Padding,
				W: newWidth,
				H: newHeight,
			}
		}

		DrawImage(renderer, p.Image.Data, imageRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	} else if p.PreviewMode == PreviewModeText {
		lines := strings.Split(p.Text, "\n")

		for index, line := range lines {
			l := strings.TrimSpace(line)
			if len(l) == 0 {
				continue
			}

			l = app.Font.ClipStringNoEllipsis(l, insetRect.W-p.Padding*2)

			lineWidth := app.Font.GetStringWidth(l)
			lineRect := sdl.Rect{
				X: insetRect.X + p.Padding,
				Y: insetRect.Y + p.Padding + app.Font.Size*int32(index),
				W: lineWidth,
				H: app.Font.Size,
			}

			if lineRect.Y+lineRect.H >= insetRect.Y+insetRect.H {
				continue
			}

			DrawText(renderer, &app.Font, l, &lineRect, GetColor(theme, "text_color"))
		}
	} else if p.PreviewMode == PreviewModeUnsupported {
		textWidth := app.Font.GetStringWidth("Preview unsupported")
		textRect := sdl.Rect{
			X: insetRect.X + (insetRect.W-textWidth)/2,
			Y: insetRect.Y + (insetRect.H-app.Font.Size)/2,
			W: textWidth,
			H: app.Font.Size,
		}
		DrawText(renderer, &app.Font, "Preview unsupported", &textRect, GetColor(theme, "text_color"))
	}
}
