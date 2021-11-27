package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func DrawText(renderer *sdl.Renderer, font *Font, text string, rect *sdl.Rect, color sdl.Color) {
	surface, err := font.Data.RenderUTF8Blended(text, color)
	checkError(err)
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	checkError(err)
	defer texture.Destroy()

	renderer.Copy(texture, nil, rect)
}

func DrawRect(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.FillRect(rect)
}

func DrawRectOutline(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	top := sdl.Rect{X: rect.X, Y: rect.Y, W: rect.W, H: 1}
	right := sdl.Rect{X: rect.X + rect.W - 1, Y: rect.Y, W: 1, H: rect.H}
	bottom := sdl.Rect{X: rect.X, Y: rect.Y + rect.H - 1, W: rect.W, H: 1}
	left := sdl.Rect{X: rect.X, Y: rect.Y, W: 1, H: rect.H}

	renderer.FillRect(&top)
	renderer.FillRect(&right)
	renderer.FillRect(&bottom)
	renderer.FillRect(&left)
}

func DrawRectTransparent(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	DrawRect(renderer, rect, color)
	renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
}

func DrawRect3D(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	DrawRect(renderer, rect, color)

	highlight := sdl.Color{R: 255, G: 255, B: 255, A: 25}
	shadow := sdl.Color{R: 0, G: 0, B: 0, A: 122}

	DrawRectHighlight(renderer, rect, highlight, 1)
	DrawRectShadow(renderer, rect, shadow, 1)
}

func DrawRect3DInset(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	DrawRect(renderer, rect, color)

	highlight := sdl.Color{R: 0, G: 0, B: 0, A: 122}
	shadow := sdl.Color{R: 255, G: 255, B: 255, A: 25}

	DrawRectHighlight(renderer, rect, highlight, 1)
	DrawRectShadow(renderer, rect, shadow, 1)
}

func DrawRectHighlight(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color, size int32) {
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	DrawRect(renderer, &sdl.Rect{X: rect.X, Y: rect.Y, W: rect.W, H: size}, color)
	DrawRect(renderer, &sdl.Rect{X: rect.X, Y: rect.Y, W: size, H: rect.H}, color)
	renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
}

func DrawRectShadow(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color, size int32) {
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	DrawRect(renderer, &sdl.Rect{X: rect.X + rect.W - size, Y: rect.Y, W: size, H: rect.H}, color)
	DrawRect(renderer, &sdl.Rect{X: rect.X, Y: rect.Y + rect.H - size, W: rect.W, H: size}, color)
	renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
}

func DrawImage(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect, color sdl.Color) {
	texture.SetColorMod(color.R, color.G, color.B)
	renderer.Copy(texture, nil, &rect)
}
