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

	// @TODO (!important) utility function to get inside border rects of rect
	// Draw top and left highlights
	DrawRect(renderer, &sdl.Rect{ // Top
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: 1,
	}, highlight)
	DrawRect(renderer, &sdl.Rect{ // Left
		X: rect.X,
		Y: rect.Y,
		W: 1,
		H: rect.H,
	}, highlight)

	// Draw right and bottom shadows
	DrawRect(renderer, &sdl.Rect{ // Right
		X: rect.X + rect.W - 1,
		Y: rect.Y,
		W: 1,
		H: rect.H,
	}, shadow)
	DrawRect(renderer, &sdl.Rect{ // Bottom
		X: rect.X,
		Y: rect.Y + rect.H - 1,
		W: rect.W,
		H: 1,
	}, shadow)
}

func DrawImage(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect, color sdl.Color) {
	texture.SetColorMod(color.R, color.G, color.B)
	renderer.Copy(texture, nil, &rect)
}
