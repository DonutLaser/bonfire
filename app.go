package main

import (
	"strings"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
)

type Mode int32

const (
	Mode_Normal Mode = iota
	Mode_Drive_Selection
	Mode_Goto
)

type App struct {
	Font Font

	Breadcrumbs Breadcrumbs
	ItemView    ItemView

	Mode Mode
}

// @TODO (!important) handle window resizing
func NewApp(windowWidth int32, windowHeight int32) (result App) {
	result.Breadcrumbs = *NewBreadcrumbs(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28})
	result.ItemView = *NewItemView(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28})
	result.Font = LoadFont("assets/fonts/consolab.ttf", 12)

	return
}

func (app *App) Tick(input *Input) {
	if app.Mode == Mode_Drive_Selection {
		app.handleInputDriveSelection(input)
		return
	}

	if app.Mode == Mode_Goto {
		app.handleInputGoto(input)
		return
	}

	app.handleInputNormal(input)
}

func (app *App) handleInputNormal(input *Input) {
	if input.Backspace {
		// @TODO (!important) fix crash when going outside from the root of the drive
		app.Breadcrumbs.Pop()
		app.ItemView.GoOutside()
		return
	}

	switch input.TypedCharacter {
	case ':':
		app.Mode = Mode_Drive_Selection
	case 'g':
		app.Mode = Mode_Goto
	case 'j':
		app.ItemView.NavigateDown()
	case 'k':
		app.ItemView.NavigateUp()
	}
}

func (app *App) handleInputDriveSelection(input *Input) {
	if input.TypedCharacter == 0 {
		return
	}

	if !unicode.IsLetter(rune(input.TypedCharacter)) {
		app.Mode = Mode_Normal
		return
	}

	var sb strings.Builder
	sb.WriteString(strings.ToUpper(string(input.TypedCharacter)))
	sb.WriteString(":")

	app.Breadcrumbs.Clear()
	app.Breadcrumbs.Push(sb.String())

	app.ItemView.ShowFolder(sb.String())

	app.Mode = Mode_Normal
}

func (app *App) handleInputGoto(input *Input) {
	switch input.TypedCharacter {
	case 'd':
		activeFolder := app.ItemView.GoInside()
		if activeFolder != "" {
			app.Breadcrumbs.Push(activeFolder)
		}

		app.Mode = Mode_Normal
	}
}

func (app *App) Render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	app.Breadcrumbs.Render(renderer, &app.Font)
	app.ItemView.Render(renderer, &app.Font)

	renderer.Present()
}
