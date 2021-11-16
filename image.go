package main

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type Image struct {
	Data   *sdl.Texture
	Width  int32
	Height int32
}

func LoadImage(path string, renderer *sdl.Renderer) (result Image) {
	image, err := img.Load(path)
	if err != nil {
		NotifyError(err.Error())
		return
	}

	texture, err := renderer.CreateTextureFromSurface(image)
	if err != nil {
		NotifyError(err.Error())
		return
	}

	result = Image{
		Data:   texture,
		Width:  image.W,
		Height: image.H,
	}

	image.Free()

	return
}

func LoadIcon(path string) *sdl.Surface {
	image, err := img.Load(path)
	if err != nil {
		NotifyError(err.Error())
		return nil
	}

	return image
}

func (image *Image) Render(renderer *sdl.Renderer, position sdl.Point, color sdl.Color) {
	rect := sdl.Rect{
		X: position.X,
		Y: position.Y,
		W: image.Width,
		H: image.Height,
	}

	DrawImage(renderer, image.Data, rect, color)
}

func (image *Image) Unload() {
	image.Data.Destroy()
}
