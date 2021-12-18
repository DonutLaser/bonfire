package main

import (
	"path"
	"strings"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
)

// @TODO (!important) copy files and folders
// @TODO have icons next to files/folders
// @TODO (!important) lazy initialize compnents that are not needed right away
// @TODO (!important) make sure changes in the directory that are made outside the app are automatically reflected in the app

type Mode int32

const (
	Mode_Normal Mode = iota
	Mode_Drive_Selection
	Mode_Goto
)

type Clipboard struct {
	Directory string
	Name      string
	Type      ItemType
}

type App struct {
	Font         Font
	FavoriteIcon Image
	Theme
	AvailableThemes []string
	AvailabelDrives []string

	ActiveView  int32
	ViewCount   int32
	WindowRects []sdl.Rect

	Breadcrumbs  []Breadcrumbs
	ItemViews    []*ItemView
	QuickOpen    QuickOpen
	Notification Notification
	InfoViews    []InfoView
	Previews     []Preview

	Mode      Mode
	LastError NotificationEvent
	Settings
	Renderer *sdl.Renderer

	NormalKeyMap map[byte]Shortcut
	Clipboard
}

func NewApp(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result *App) {
	result = &App{}

	result.Font = LoadFont("assets/fonts/consolab.ttf", 12)
	result.FavoriteIcon = LoadImage("assets/images/favorite.png", renderer)
	result.AvailableThemes = GetAvailableThemes()
	result.AvailabelDrives = GetAvailableDrives()

	result.ActiveView = 0
	result.ViewCount = 1
	result.WindowRects = []sdl.Rect{{X: 0, Y: 0, W: windowWidth, H: windowHeight}}

	result.Breadcrumbs = []Breadcrumbs{*NewBreadcrumbs(sdl.Rect{X: 0, Y: 0, W: windowWidth, H: 28}, result.AvailabelDrives)}
	result.ItemViews = []*ItemView{NewItemView(sdl.Rect{X: 0, Y: 28, W: windowWidth, H: windowHeight - 28}, result)}
	// Only the width matters here, because the position is relative to parent component and height is dynamic
	result.QuickOpen = *NewQuickOpen(sdl.Rect{X: 0, Y: 0, W: 394, H: 0})
	result.Notification = *NewNotification()
	result.InfoViews = []InfoView{*NewInfoView()}
	result.Previews = []Preview{*NewPreview()}

	result.GoToDrive('D')
	result.Mode = Mode_Normal
	result.Settings = NewSettings()
	result.Renderer = renderer

	result.Theme = *LoadTheme(result.Settings.ThemeName)

	result.ItemViews[result.ActiveView].SetFavorites(result.Settings.Favorites)

	globalNotificationHandler = func(e NotificationEvent) {
		result.ShowNotification(e)
	}

	result.NormalKeyMap = map[byte]Shortcut{}
	result.NormalKeyMap[':'] = Shortcut{Ctrl: false, Alt: false, Callback: func() {
		result.Mode = Mode_Drive_Selection
		result.Breadcrumbs[result.ActiveView].ShowAvailableDrives(true)
	}}
	result.NormalKeyMap['e'] = Shortcut{Ctrl: false, Alt: true, Callback: func() {
		result.ShowNotification(result.LastError)
	}}
	result.NormalKeyMap['.'] = Shortcut{Ctrl: true, Alt: true, Callback: func() {
		result.SelectTheme(result.AvailableThemes)
	}}
	result.NormalKeyMap['}'] = Shortcut{Ctrl: true, Alt: false, Callback: func() {
		result.AddView()
	}}
	result.NormalKeyMap['{'] = Shortcut{Ctrl: true, Alt: false, Callback: func() {
		result.RemoveView()
	}}
	result.NormalKeyMap[']'] = Shortcut{Ctrl: true, Alt: false, Callback: func() {
		result.GoToNextView()
	}}
	result.NormalKeyMap['['] = Shortcut{Ctrl: true, Alt: false, Callback: func() {
		result.GoToPrevView()
	}}

	return
}

func (app *App) Close() {
	// @TODO (!important) What if the program is closed in such a way that Save function is not called?
	app.Settings.Save(false)
	app.Font.Unload()
	app.FavoriteIcon.Unload()
}

func (app *App) GetIcon() *sdl.Surface {
	return LoadIcon("assets/images/icon.png")
}

func (app *App) Resize(windowWidth int32, windowHeight int32) {
	singleWidth := windowWidth / app.ViewCount

	for i := int32(0); i < app.ViewCount; i++ {
		app.WindowRects[i].W = singleWidth
		app.WindowRects[i].H = windowHeight

		app.Breadcrumbs[i].Resize(sdl.Rect{X: singleWidth * i, Y: 0, W: singleWidth, H: 28})
		app.ItemViews[i].Resize(sdl.Rect{X: singleWidth * i, Y: 28, W: singleWidth, H: windowHeight - 28})
	}

	for index, bc := range app.Breadcrumbs {
		bc.Resize(sdl.Rect{X: singleWidth * int32(index), Y: 0, W: singleWidth, H: 28})
	}
}

func (app *App) Tick(input *Input) {
	if app.Notification.IsOpen {
		app.Notification.Tick()
	}

	if app.InfoViews[app.ActiveView].IsOpen {
		app.InfoViews[app.ActiveView].Tick(input)
	}

	if app.Previews[app.ActiveView].IsOpen {
		app.Previews[app.ActiveView].Tick(input)
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
	// @TODO (!important) p to paste a folder
	// @TODO (!important) P to paste an item contents (files if copying folder), shouldn't do anything for copied files

	app.ItemViews[app.ActiveView].Tick(input)

	if app.ItemViews[app.ActiveView].ConsumingInput {
		return
	}

	shortcut, ok := app.NormalKeyMap[input.TypedCharacter]
	if ok {
		if input.Ctrl == shortcut.Ctrl && input.Alt == shortcut.Alt {
			shortcut.Callback()
		}
	}
}

func (app *App) handleInputDriveSelection(input *Input) {
	if input.Escape {
		app.Mode = Mode_Normal
		app.Breadcrumbs[app.ActiveView].ShowAvailableDrives(false)
		return
	}

	if input.TypedCharacter == 0 {
		return
	}

	if !unicode.IsLetter(rune(input.TypedCharacter)) || input.Escape {
		app.Mode = Mode_Normal
		return
	}

	app.GoToDrive(input.TypedCharacter)
	app.Mode = Mode_Normal
	app.Breadcrumbs[app.ActiveView].ShowAvailableDrives(false)
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
	success := app.ItemViews[app.ActiveView].ShowFolder(withSlash)

	if success {
		app.Breadcrumbs[app.ActiveView].Clear()
		app.Breadcrumbs[app.ActiveView].Push(withoutSlash)
	}
}

func (app *App) SelectFavorite(favorites []string) {
	app.QuickOpen.Open(favorites, func(favorite string) {
		app.ItemViews[app.ActiveView].OpenFavorite(favorite)
	})
}

func (app *App) SelectTheme(themes []string) {
	// @TODO (!important) should show preview when hovering over a theme
	app.QuickOpen.Open(themes, func(theme string) {
		app.Theme = *LoadTheme(theme)
		app.Settings.SetTheme(theme)
	})
}

func (app *App) FindInCurrentFolder(items []string) {
	app.QuickOpen.Open(items, func(item string) {
		app.ItemViews[app.ActiveView].OpenItem(item)
	})
}

func (app *App) ShowNotification(event NotificationEvent) {
	app.Notification.Show(event)

	if event.Type == NotificationError {
		app.LastError = event
	}
}

func (app *App) ShowFileInfo(info Info) {
	app.InfoViews[app.ActiveView].Show(info)
}

func (app *App) ShowPreview(directory string, name string) {
	fileType := GetFileType(name)

	switch fileType {
	case FileTypeImage:
		image := LoadImage(path.Join(directory, name), app.Renderer)
		app.Previews[app.ActiveView].ShowImage(name, &image)
	case FileTypeText:
		text := ReadFile(path.Join(directory, name))
		app.Previews[app.ActiveView].ShowText(name, text)
	default:
		app.Previews[app.ActiveView].ShowPreviewUnsupported(name)
	}
}

// Used when the size is calculated in another thread
func (app *App) SetFileInfoSize(size string) {
	app.InfoViews[app.ActiveView].Info.Size = size
}

func (app *App) AddView() {
	if app.ViewCount == 3 {
		return
	}

	newCount := app.ViewCount + 1

	fullRect := sdl.Rect{X: 0, Y: 0, W: 0, H: app.WindowRects[0].H}
	for i := int32(0); i < app.ViewCount; i++ {
		fullRect.W += app.WindowRects[i].W
	}
	singleWidth := fullRect.W / newCount

	for i := int32(0); i < app.ViewCount; i++ {
		app.WindowRects[i] = sdl.Rect{X: singleWidth * i, Y: 0, W: singleWidth, H: app.WindowRects[0].H}
	}
	app.WindowRects = append(app.WindowRects, sdl.Rect{X: singleWidth * app.ViewCount, Y: 0, W: singleWidth, H: app.WindowRects[0].H})

	for i := int32(0); i < app.ViewCount; i++ {
		app.Breadcrumbs[i].Resize(sdl.Rect{X: singleWidth * i, Y: 0, W: singleWidth, H: 28})
		app.ItemViews[i].Resize(sdl.Rect{X: singleWidth * i, Y: 28, W: singleWidth, H: app.WindowRects[0].H - 28})
	}
	app.Breadcrumbs = append(app.Breadcrumbs, *NewBreadcrumbs(sdl.Rect{X: singleWidth * app.ViewCount, Y: 0, W: singleWidth, H: 28}, app.AvailabelDrives))
	app.ItemViews = append(app.ItemViews, NewItemView(sdl.Rect{X: singleWidth * app.ViewCount, Y: 28, W: singleWidth, H: app.WindowRects[0].H - 28}, app))
	app.InfoViews = append(app.InfoViews, *NewInfoView())
	app.Previews = append(app.Previews, *NewPreview())

	app.ActiveView = newCount - 1
	app.ViewCount = newCount

	app.ItemViews[app.ActiveView].SetFavorites(app.Settings.Favorites)
	app.GoToDrive('D')
}

func (app *App) RemoveView() {
	if app.ViewCount == 1 {
		return
	}

	newCount := app.ViewCount - 1

	fullRect := sdl.Rect{X: 0, Y: 0, W: 0, H: app.WindowRects[0].H}
	for i := int32(0); i < app.ViewCount; i++ {
		fullRect.W += app.WindowRects[i].W
	}
	singleWidth := fullRect.W / newCount

	for i := int32(0); i < newCount; i++ {
		app.WindowRects[i] = sdl.Rect{X: singleWidth * i, Y: 0, W: singleWidth, H: app.WindowRects[0].H}
	}
	app.WindowRects = app.WindowRects[:newCount]

	for i := int32(0); i < newCount; i++ {
		app.Breadcrumbs[i].Resize(sdl.Rect{X: singleWidth * i, Y: 0, W: singleWidth, H: 28})
		app.ItemViews[i].Resize(sdl.Rect{X: singleWidth * i, Y: 28, W: singleWidth, H: app.WindowRects[0].H - 28})
	}
	app.Breadcrumbs = app.Breadcrumbs[:newCount]
	app.ItemViews = app.ItemViews[:newCount]
	app.InfoViews = app.InfoViews[:newCount]
	app.Previews = app.Previews[:newCount]

	if app.ActiveView >= newCount {
		app.ActiveView = newCount - 1
	}
	app.ViewCount = newCount
}

func (app *App) GoToNextView() {
	if app.ActiveView < app.ViewCount-1 {
		app.ActiveView++
	}
}

func (app *App) GoToPrevView() {
	if app.ActiveView > 0 {
		app.ActiveView--
	}
}

func (app *App) Copy(name string, directory string, itemType ItemType) {
	app.Clipboard.Name = name
	app.Clipboard.Directory = directory
	app.Type = itemType
}

func (app *App) MoveItemToNextView(name string, directory string, itemType ItemType) {
	nextView := app.ActiveView + 1
	if nextView >= app.ViewCount {
		return
	}

	app.ActiveView = nextView

	app.Copy(name, directory, itemType)
	app.ItemViews[app.ActiveView].Paste()
}

func (app *App) MoveItemToPrevView(name string, directory string, itemType ItemType) {
	prevView := app.ActiveView - 1
	if prevView < 0 {
		return
	}

	app.ActiveView = prevView

	app.Copy(name, directory, itemType)
	app.ItemViews[app.ActiveView].Paste()
}

func (app *App) GetClipboard() *Clipboard {
	return &app.Clipboard
}

func (app *App) Render() {
	app.Renderer.SetDrawColor(0, 0, 0, 255)
	app.Renderer.Clear()

	for i := int32(0); i < app.ViewCount; i++ {
		app.Breadcrumbs[i].Render(app.Renderer, &app.Font, app.Theme.BreadcrumbsTheme)
		app.ItemViews[i].Render(app.Renderer, app, app.ActiveView == i)
	}

	if app.Mode == Mode_Drive_Selection {
		rect := sdl.Rect{X: app.Breadcrumbs[app.ActiveView].Rect.X, Y: app.Breadcrumbs[app.ActiveView].Rect.H, W: app.WindowRects[app.ActiveView].W, H: app.ItemViews[app.ActiveView].Rect.H}
		DrawRectTransparent(app.Renderer, &rect, sdl.Color{R: 0, G: 0, B: 0, A: 150})
	}

	if app.Notification.IsOpen {
		fullRect := sdl.Rect{X: 0, Y: 0, W: app.WindowRects[0].W * app.ViewCount, H: app.WindowRects[0].H}
		app.Notification.Render(app.Renderer, &fullRect, app)
	}

	for i := int32(0); i < app.ViewCount; i++ {
		if app.InfoViews[i].IsOpen {
			app.InfoViews[i].Render(app.Renderer, &app.WindowRects[i], app)
		}

		if app.Previews[i].IsOpen {
			app.Previews[i].Render(app.Renderer, &app.WindowRects[i], app)
		}
	}

	if app.QuickOpen.IsOpen {
		DrawRectTransparent(app.Renderer, &app.WindowRects[app.ActiveView], sdl.Color{R: 0, G: 0, B: 0, A: 150})
		app.QuickOpen.Render(app.Renderer, &app.ItemViews[app.ActiveView].Rect, &app.Font, app.Theme.QuickOpenTheme, app.Theme.InputFieldTheme)
	}

	app.Renderer.Present()
}
