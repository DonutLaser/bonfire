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

func DrawRectTransparent(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.FillRect(rect)

	renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE)
}

func DrawRect3D(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	DrawRect(renderer, rect, color)

	// @TODO (!important) use theme
	highlight := sdl.Color{
		R: 44,
		G: 50,
		B: 61,
		A: 255,
	}

	shadow := sdl.Color{
		R: 10,
		G: 13,
		B: 19,
		A: 255,
	}

	DrawRectHighlight(renderer, rect, highlight, 1)
	DrawRectShadow(renderer, rect, shadow, 1)
}

func DrawRectHighlight(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color, size int32) {
	DrawRect(renderer, &sdl.Rect{ // Top
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: size,
	}, color)
	DrawRect(renderer, &sdl.Rect{ // Left
		X: rect.X,
		Y: rect.Y,
		W: size,
		H: rect.H,
	}, color)
}

func DrawRectShadow(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color, size int32) {
	DrawRect(renderer, &sdl.Rect{ // Right
		X: rect.X + rect.W - size,
		Y: rect.Y,
		W: size,
		H: rect.H,
	}, color)
	DrawRect(renderer, &sdl.Rect{ // Bottom
		X: rect.X,
		Y: rect.Y + rect.H - size,
		W: rect.W,
		H: size,
	}, color)
}

func DrawImage(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect, color sdl.Color) {
	texture.SetColorMod(color.R, color.G, color.B)
	renderer.Copy(texture, nil, &rect)
}
