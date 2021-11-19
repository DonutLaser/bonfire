package main

import (
	"strings"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
)

// @TODO (!important) open favorite file
// @TODO (!important) copy files and folders
// @TODO (!important) advanced rename files and folders
// @TODO (!important) show/hide hidden files
// @TODO (!important) custom themes
// @TODO (!important) open files
// @TODO (!important) tag files
// @TODO have icons next to files/folders
// @TODO (!important) lazy initialize compnents that are not needed right away
// @TODO (!important) make sure changes in the directory that are made outside the app are automatically reflected in the app

type Mode int32

const (
	Mode_Normal Mode = iota
	Mode_Drive_Selection
	Mode_Goto
)

type App struct {
	Font Font
	Theme
	WindowRect sdl.Rect

	Breadcrumbs
	ItemView
	QuickOpen
	Notification

	Mode      Mode
	LastError NotificationEvent
}

func NewApp(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result *App) {
	result = &App{}

	result.Font = LoadFont("assets/fonts/consolab.ttf", 12)
	result.Theme = *LoadTheme("default")
	result.WindowRect = sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}

	favoriteIcon := LoadImage("assets/images/favorite.png", renderer)

	result.Breadcrumbs = *NewBreadcrumbs(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28})
	result.ItemView = *NewItemView(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28}, &favoriteIcon, result)
	// Only the width matters here, because the position is relative to parent component and height is dynamic
	result.QuickOpen = *NewQuickOpen(sdl.Rect{X: 0, Y: 0, W: 394, H: 0})
	result.Notification = *NewNotification()

	result.GoToDrive('D')
	result.Mode = Mode_Normal

	globalNotificationHandler = func(e NotificationEvent) {
		result.ShowNotification(e)
	}

	return
}

func (app *App) Close() {
	app.Font.Unload()
}

func (app *App) Resize(windowWidth int32, windowHeight int32) {
	app.WindowRect.W = windowWidth
	app.WindowRect.H = windowHeight
	app.Breadcrumbs.Resize(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28})
	app.ItemView.Resize(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28})
}

func (app *App) Tick(input *Input) {
	if app.Notification.IsOpen {
		app.Notification.Tick()
	}

	if app.QuickOpen.IsOpen {
		app.QuickOpen.Tick(input)
		return
	}

	if app.Mode == Mode_Drive_Selection {
		app.handleInputDriveSelection(input)
		return
	}

	app.handleInputNormal(input)
}

func (app *App) handleInputNormal(input *Input) {
	// @TODO (!important) R to advanded rename a group of selected items
	// @TODO (!important) y to copy an item
	// @TODO (!important) p to paste an item
	// @TODO (!important) P to paste an item contents (files if copying folder or file contents if copying a file)
	// @TODO (!important) ctrl + h to toggle visibility of hidden items
	// @TODO (!important) c to change an extension
	// @TODO (!important) M to mark the parent folder as favorite
	// @TODO (!important) ctrl + shift + o to open any file or folder by writing a full path
	// @TODO (!important) X on a folder to take files out of the folder and remove only the folder and leave the files intact
	// @TODO (!important) D to duplicate a file
	// @TODO (!important) ctrl + x on a folder to force remove it and all its contents

	app.ItemView.Tick(input)

	if app.ItemView.ConsumingInput {
		return
	}

	switch input.TypedCharacter {
	case ':':
		app.Mode = Mode_Drive_Selection
	case 'e':
		if input.Alt {
			app.ShowNotification(app.LastError)
		}
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

func (app *App) GoToDrive(drive byte) {
	var sb strings.Builder
	sb.WriteString(strings.ToUpper(string(drive)))
	sb.WriteString(":")

	withoutSlash := sb.String()

	sb.WriteString("/")
	withSlash := sb.String()

	// When we are opening the drive where the cwd is, go for some reason reads the cwd, not the drive.
	// Adding a slash after the colon seems to fix this for whatever reason.
	success := app.ItemView.ShowFolder(withSlash)

	if success {
		app.Breadcrumbs.Clear()
		app.Breadcrumbs.Push(withoutSlash)
	}
}

func (app *App) GoToDirectory(fullPath string) {
	app.Breadcrumbs.Set(fullPath)
	app.ItemView.ShowFolder(fullPath)
}

func (app *App) SelectFavorite(favorites []string) {
	app.QuickOpen.Open(favorites, func(favorite string) {
		app.GoToDirectory(favorite)
	})
}

func (app *App) FindInCurrentFolder(items []string) {
	app.QuickOpen.Open(items, func(item string) {
		app.ItemView.OpenFolder(item)
	})
}

func (app *App) ShowNotification(event NotificationEvent) {
	app.Notification.Show(event)

	if event.Type == NotificationError {
		app.LastError = event
	}
}

func (app *App) Render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	app.Breadcrumbs.Render(renderer, &app.Font, app.Theme.BreadcrumbsTheme)
	app.ItemView.Render(renderer, app)

	if app.Notification.IsOpen {
		app.Notification.Render(renderer, app)
	}

	if app.QuickOpen.IsOpen {
		DrawRectTransparent(renderer, &app.WindowRect, sdl.Color{R: 0, G: 0, B: 0, A: 150})
		app.QuickOpen.Render(renderer, &app.ItemView.Rect, &app.Font, app.Theme.QuickOpenTheme, app.Theme.InputFieldTheme)
	}

	renderer.Present()
}
