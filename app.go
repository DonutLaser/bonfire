package main

import (
	"strings"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
)

// @TODO (!important) marking folders and files as favorite
// @TODO (!important) open favorite files and folders
// @TODO (!important) copy files and folders
// @TODO (!important) group files/folders into a new folder
// @TODO (!important) search files in the current folder
// @TODO (!important) remove files and folders
// @TODO (!important) select multiple files/folders
// @TODO (!important) simple rename files and folders
// @TODO (!important) advanced rename files and folders
// @TODO (!important) show/hide hidden files
// @TODO (!important) create new file
// @TODO (!important) create new folder
// @TODO (!important) custom themes
// @TODO (!important) open files
// @TODO (!important) tag files
// @TODO (!important) w: opens this project folder immediatelly
// @TODO have icons next to files/folders

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

func NewApp(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result App) {
	result.Font = LoadFont("assets/fonts/consolab.ttf", 12)
	favoriteIcon := LoadImage("assets/images/favorite.png", renderer)

	result.Breadcrumbs = *NewBreadcrumbs(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28})
	result.ItemView = *NewItemView(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28}, &favoriteIcon)

	result.GoToDrive('D')
	result.Mode = Mode_Normal

	return
}

func (app *App) Close() {
	app.Font.Unload()
}

func (app *App) Resize(windowWidth int32, windowHeight int32) {
	app.Breadcrumbs.Resize(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28})
	app.ItemView.Resize(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28})
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
	// @TODO (!important) ctrl + alt + p to open one of the favorite items
	// @TODO (!important) ctrl + p to open one of the items in the current folder
	// @TODO (!important) m to mark item as favorite
	// @TODO (!important) c to rename an item
	// @TODO (!important) C to advanded rename a group of selected items
	// @TODO (!important) v to select a single item
	// @TODO (!important) V to start group selection
	// @TODO (!important) y to copy an item
	// @TODO (!important) p to paste an item
	// @TODO (!important) P to paste an item contents (files if copying folder or file contents if copying a file)
	// @TODO (!important) x to delete an item
	// @TODO (!important) / to search for an item in the current folder
	// @TODO (!important) i to create a new file
	// @TODO (!important) I to create a new folder
	// @TODO (!important) ctrl + g to put selected items into a new folder
	// @TODO (!important) ctrl + h to toggle visibility of hidden items
	// @TODO (!important) r to change an extension

	if input.Backspace {
		// @TODO (!important) fix crash when going outside from the root of the drive
		// @TODO (!important) this should retain the last position so that when you go back, the active item doesn't always become 0
		app.Breadcrumbs.Pop()
		app.ItemView.GoOutside()
		return
	}

	switch input.TypedCharacter {
	case ':':
		app.Mode = Mode_Drive_Selection
	case 'g':
		app.Mode = Mode_Goto
	case 'h':
		app.ItemView.NavigateLeft()
	case 'j':
		app.ItemView.NavigateDown()
	case 'k':
		app.ItemView.NavigateUp()
	case 'l':
		app.ItemView.NavigateRight()
	case 'G':
		app.ItemView.NavigateLastInColumn()
	case 'm':
		app.ItemView.MarkActiveAsFavorite()
	}
}

func (app *App) handleInputDriveSelection(input *Input) {
	if input.TypedCharacter == 0 {
		return
	}

	if !unicode.IsLetter(rune(input.TypedCharacter)) || input.Escape {
		app.Mode = Mode_Normal
		return
	}

	app.GoToDrive(input.TypedCharacter)
	app.Mode = Mode_Normal
}

func (app *App) handleInputGoto(input *Input) {
	if input.TypedCharacter == 0 {
		return
	}

	if input.Escape {
		app.Mode = Mode_Normal
		return
	}

	switch input.TypedCharacter {
	case 'd':
		activeFolder := app.ItemView.GoInside()
		if activeFolder != "" {
			app.Breadcrumbs.Push(activeFolder)
		}

		app.Mode = Mode_Normal
	case 'g':
		app.ItemView.NavigateFirstInColumn()
		app.Mode = Mode_Normal
	default:
		app.Mode = Mode_Normal
	}
}

func (app *App) GoToDrive(drive byte) {
	var sb strings.Builder
	sb.WriteString(strings.ToUpper(string(drive)))
	sb.WriteString(":")

	app.Breadcrumbs.Clear()
	app.Breadcrumbs.Push(sb.String())

	app.ItemView.ShowFolder(sb.String())
}

func (app *App) Render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	app.Breadcrumbs.Render(renderer, &app.Font)
	app.ItemView.Render(renderer, &app.Font)

	renderer.Present()
}
