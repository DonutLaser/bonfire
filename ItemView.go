package main

import (
	"os"
	"path"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/veandco/go-sdl2/sdl"
)

type ItemType int32

const (
	ItemTypeFile ItemType = iota
	ItemTypeFolder
)

type Clipboard struct {
	Directory string
	Name      string
	Type      ItemType
}

type Favorite struct {
	FullPath string
	Type     ItemType
}

type Item struct {
	Type     ItemType
	FileType FileType
	Name     string

	IsSelected       bool
	IsFavorite       bool
	RenameInProgress bool
}

type ItemView struct {
	Items        []Item
	ActiveItem   int32
	ActiveColumn int32
	CurrentPath  string

	Favorites []Favorite
	Clipboard

	Rect              sdl.Rect
	ItemHeight        int32
	ItemWidth         int32
	MaxItemsPerColumn int32
	Columns           int32

	App  *App
	Mode Mode

	FavoriteIcon   *Image
	Input          *InlineInputField
	ConsumingInput bool
	SelectionMode  bool
	SelectionStart int32
}

func NewItemView(rect sdl.Rect, favoriteIcon *Image, app *App) *ItemView {
	return &ItemView{
		ActiveItem:        -1,
		ActiveColumn:      0,
		Rect:              rect,
		ItemHeight:        24,
		ItemWidth:         394,
		MaxItemsPerColumn: rect.H / 24,
		App:               app,
		Mode:              Mode_Normal,
		FavoriteIcon:      favoriteIcon,
		Input:             NewInlineInputField(),
	}
}

func (iv *ItemView) Close() {
	iv.FavoriteIcon.Unload()
}

func (iv *ItemView) SetActiveByName(name string) {
	for index, item := range iv.Items {
		if item.Name == name {
			iv.ActiveItem = int32(index)
			iv.ActiveColumn = int32(index) / iv.MaxItemsPerColumn
			break
		}
	}
}

func (iv *ItemView) SetFavorites(favorites []string) {
	iv.Favorites = make([]Favorite, 0)

	for _, favorite := range favorites {
		iv.Favorites = append(iv.Favorites, Favorite{
			FullPath: favorite,
			Type:     GetFileType(favorite),
		})
	}

	for index, item := range iv.Items {
		iv.Items[index].IsFavorite = iv.favoriteIndex(path.Join(iv.CurrentPath, item.Name)) >= 0
	}
}

func (iv *ItemView) ShowFolder(fullPath string) bool {
	items, success := ReadDirectory(fullPath)
	if !success {
		return false
	}

	var files []Item
	var folders []Item

	for _, file := range items {
		if file.IsDir() {
			folders = append(folders, Item{
				Type: ItemTypeFolder,
				Name: file.Name(),
			})
		} else {
			files = append(files, Item{
				Type:     ItemTypeFile,
				Name:     file.Name(),
				FileType: FileType(iv.getFileType(file.Name())),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	sort.Slice(folders, func(i, j int) bool {
		return strings.ToLower(folders[i].Name) < strings.ToLower(folders[j].Name)
	})

	iv.Items = make([]Item, 0)
	iv.Items = append(iv.Items, folders...)
	iv.Items = append(iv.Items, files...)

	for index, item := range iv.Items {
		iv.Items[index].IsFavorite = iv.favoriteIndex(path.Join(fullPath, item.Name)) >= 0
	}

	iv.Columns = int32(len(iv.Items)) / iv.MaxItemsPerColumn
	iv.Columns++ // Basically ceiling the number

	iv.ActiveItem = 0
	iv.CurrentPath = fullPath

	return true
}

func (iv *ItemView) GetActiveFileInfo() (result Info) {
	item := iv.Items[iv.ActiveItem]

	fullPath := path.Join(iv.CurrentPath, item.Name)
	stats, err := os.Stat(fullPath)
	if err != nil {
		NotifyError(err.Error())
	}

	result.Name = iv.Items[iv.ActiveItem].Name

	if item.Type == ItemTypeFolder {
		result.Size = "Calculating..."

		// Getting the directory size might take a lot of time, therefore, we do it in a goroutine to prevent blocking the interface.
		// When the directory size is finally calculated, it will be sent directly to the InfoView.
		go func() {
			size := bytesToString(GetDirectorySize(path.Join(iv.CurrentPath, item.Name)))
			iv.App.SetFileInfoSize(size)
		}()
	} else {
		result.Size = bytesToString(stats.Size())
	}

	result.Created = time.Unix(0, stats.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds()).Format("2006-01-02 15:05:05")
	result.Modified = stats.ModTime().Format("2006-01-02 15:04:05")

	return
}

func (iv *ItemView) OpenPath(fullPath string) {
	index := iv.favoriteIndex(fullPath)
	if index < 0 {
		return
	}

	favorite := iv.Favorites[index]
	if favorite.Type == ItemTypeFolder {
		iv.App.Breadcrumbs.Set(fullPath)
		iv.ShowFolder(fullPath)
	} else if favorite.Type == ItemTypeFile {
		p := path.Dir(fullPath)

		if strings.HasSuffix(p, ":") {
			p += "/"
		}

		iv.App.Breadcrumbs.Set(p)
		iv.ShowFolder(p)

		base := path.Base(fullPath)

		iv.SetActiveByName(base)
		iv.OpenFile(base)
	}
}

func (iv *ItemView) OpenItem(name string) {
	iv.SetActiveByName(name)

	if iv.Items[iv.ActiveItem].Type == ItemTypeFolder {
		iv.OpenFolder(name)
	} else if iv.Items[iv.ActiveItem].Type == ItemTypeFile {
		iv.OpenFile(name)
	}
}

func (iv *ItemView) OpenFolder(name string) {
	iv.SetActiveByName(name)

	activeFolder := iv.GoInside()
	if activeFolder != "" {
		iv.App.Breadcrumbs.Push(activeFolder)
	}
}

func (iv *ItemView) OpenFile(name string) {
	open.Start(path.Join(iv.CurrentPath, name))
}

func (iv *ItemView) NavigateDown() {
	if iv.ActiveItem < int32(len(iv.Items)-1) && iv.ActiveItem < iv.MaxItemsPerColumn*(iv.ActiveColumn+1)-1 {
		iv.ActiveItem++
		iv.updateSelection()
	}
}

func (iv *ItemView) NavigateUp() {
	if iv.ActiveItem > iv.MaxItemsPerColumn*iv.ActiveColumn {
		iv.ActiveItem--
		iv.updateSelection()
	}
}

func (iv *ItemView) NavigateRight() {
	if iv.ActiveColumn >= iv.Columns-1 {
		return
	}

	iv.ActiveColumn++

	if iv.ActiveItem+iv.MaxItemsPerColumn >= int32(len(iv.Items)) {
		iv.ActiveItem = int32(len(iv.Items) - 1)
	} else {
		iv.ActiveItem += iv.MaxItemsPerColumn
	}

	iv.updateSelection()
}

func (iv *ItemView) NavigateLeft() {
	if iv.ActiveColumn <= 0 {
		return
	}

	iv.ActiveColumn--
	iv.ActiveItem -= iv.MaxItemsPerColumn
	iv.updateSelection()
}

func (iv *ItemView) NavigateFirstInColumn() {
	iv.ActiveItem = iv.ActiveColumn * iv.MaxItemsPerColumn
	iv.updateSelection()
}

func (iv *ItemView) NavigateLastInColumn() {
	iv.ActiveItem = iv.ActiveColumn*iv.MaxItemsPerColumn + iv.MaxItemsPerColumn - 1
	if iv.ActiveItem >= int32(len(iv.Items)) {
		iv.ActiveItem = int32(len(iv.Items) - 1)
	}

	iv.updateSelection()
}

func (iv *ItemView) GoInside() (result string) {
	if iv.ActiveItem < 0 || iv.ActiveItem >= int32(len(iv.Items)) {
		return
	}

	item := iv.Items[iv.ActiveItem]

	if item.Type == ItemTypeFolder {
		iv.ShowFolder(path.Join(iv.CurrentPath, item.Name))
		result = item.Name
	}

	return
}

func (iv *ItemView) GoOutside() {
	split := strings.Split(iv.CurrentPath, "/")
	lastName := split[len(split)-1]
	iv.CurrentPath = strings.Join(split[:len(split)-1], "/")

	if strings.HasSuffix(iv.CurrentPath, ":") {
		iv.CurrentPath += "/"
	}

	iv.ShowFolder(iv.CurrentPath)
	iv.SetActiveByName(lastName)
}

func (iv *ItemView) MarkActiveAsFavorite() {
	if iv.Items[iv.ActiveItem].IsFavorite {
		iv.Items[iv.ActiveItem].IsFavorite = false

		fullPath := path.Join(iv.CurrentPath, iv.Items[iv.ActiveItem].Name)
		iv.RemoveFavorite(fullPath)
		iv.App.Settings.RemoveFavorite(fullPath)
	} else {
		item := iv.Items[iv.ActiveItem]
		iv.Items[iv.ActiveItem].IsFavorite = true

		fullPath := path.Join(iv.CurrentPath, item.Name)
		iv.Favorites = append(iv.Favorites, Favorite{FullPath: fullPath, Type: item.Type})
		iv.App.Settings.AddFavorite(fullPath)
	}
}

func (iv *ItemView) DeleteActive() {
	if iv.ActiveItem < 0 || iv.ActiveItem >= int32(len(iv.Items)) {
		return
	}

	err := os.Remove(path.Join(iv.CurrentPath, iv.Items[iv.ActiveItem].Name))
	if err != nil {
		NotifyError(err.Error())
		return
	}

	lastActive := iv.ActiveItem

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)

	iv.ActiveItem = lastActive
	if iv.ActiveItem >= int32(len(iv.Items)) && len(iv.Items) > 0 {
		iv.ActiveItem = int32(len(iv.Items)) - 1
	} else if len(iv.Items) == 0 {
		iv.ActiveItem = -1
	}
}

func (iv *ItemView) DeleteSelected() {
	for i := 0; i < len(iv.Items); i++ {
		if !iv.Items[i].IsSelected {
			continue
		}

		err := os.Remove(path.Join(iv.CurrentPath, iv.Items[i].Name))
		if err != nil {
			NotifyError(err.Error())
			continue
		}
	}

	iv.SelectionMode = false

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)
}

func (iv *ItemView) CopyActive(showNotification bool) {
	iv.Clipboard.Name = iv.Items[iv.ActiveItem].Name
	iv.Clipboard.Directory = iv.CurrentPath
	iv.Clipboard.Type = iv.Items[iv.ActiveItem].Type

	if showNotification {
		NotifyInfo("Copied " + path.Join(iv.Clipboard.Directory, iv.Clipboard.Name))
	}
}

func (iv *ItemView) Paste() {
	if iv.Clipboard.Type == ItemTypeFolder {
		// @TODO (!important) implement pasting a folder
	} else if iv.Clipboard.Type == ItemTypeFile {
		success, name := DuplicateFile(iv.Clipboard.Directory, iv.Clipboard.Name)
		if !success {
			return
		}

		iv.ShowFolder(iv.CurrentPath)
		iv.SetActiveByName(name)
	}
}

func (iv *ItemView) DuplicateActive() {
	if iv.Items[iv.ActiveItem].Type == ItemTypeFolder {

	} else if iv.Items[iv.ActiveItem].Type == ItemTypeFile {
		iv.CopyActive(false)
		iv.Paste()
	}
}

func (iv *ItemView) RenameActive() {
	iv.Items[iv.ActiveItem].RenameInProgress = true
	iv.ConsumingInput = true

	iv.Input.Open(iv.Items[iv.ActiveItem].Name, func(value string) {
		oldName := iv.Items[iv.ActiveItem].Name
		iv.Items[iv.ActiveItem].Name = value
		iv.Items[iv.ActiveItem].RenameInProgress = false
		iv.ConsumingInput = false

		err := os.Rename(path.Join(iv.CurrentPath, oldName), path.Join(iv.CurrentPath, value))
		if err != nil {
			NotifyError(err.Error())
			iv.Items[iv.ActiveItem].Name = oldName
		}
	}, func() {
		iv.Items[iv.ActiveItem].RenameInProgress = false
		iv.ConsumingInput = false
	})
}

func (iv *ItemView) SelectActive() {
	if iv.ActiveItem < 0 || iv.ActiveItem >= int32(len(iv.Items)) {
		return
	}

	iv.Items[iv.ActiveItem].IsSelected = !iv.Items[iv.ActiveItem].IsSelected
}

func (iv *ItemView) StartSelection() {
	iv.SelectionMode = true

	iv.SelectActive()
	iv.SelectionStart = iv.ActiveItem
}

func (iv *ItemView) SelectAll(sel bool) {
	for i := 0; i < len(iv.Items); i++ {
		iv.Items[i].IsSelected = sel
	}
}

func (iv *ItemView) updateSelection() {
	if !iv.SelectionMode {
		return
	}

	from := iv.SelectionStart
	to := iv.ActiveItem
	if iv.SelectionStart > iv.ActiveItem {
		from = iv.ActiveItem
		to = iv.SelectionStart
	}

	for i := 0; i < len(iv.Items); i++ {
		iv.Items[i].IsSelected = int32(i) >= from && int32(i) <= to
	}
}

func (iv *ItemView) getSelectedItemsCount() (result int32) {
	for i := 0; i < len(iv.Items); i++ {
		if iv.Items[i].IsSelected {
			result++
		}
	}

	return
}

func (iv *ItemView) itemsToNames() []string {
	result := make([]string, len(iv.Items))

	for index, item := range iv.Items {
		result[index] = item.Name
	}

	return result
}

func (iv *ItemView) favoritesToPaths() []string {
	result := make([]string, len(iv.Favorites))

	for index, item := range iv.Favorites {
		result[index] = item.FullPath
	}

	return result
}

func (iv *ItemView) getFileType(filename string) int32 {
	if strings.HasSuffix(filename, ".exe") {
		return FileType_Exe
	}

	imageExtensions := []string{".png", ".jpg", ".jpeg", ".bmp", ".gif", ".ico"}
	for _, ext := range imageExtensions {
		if strings.HasSuffix(filename, ext) {
			return FileType_Image
		}
	}

	return FileType_Default
}

// func (iv *ItemView) getSelectedItems() (result []string) {
// 	result = make([]string, iv.getSelectedItemsCount())
// 	index := 0

// 	for i := 0; i < len(iv.Items); i++ {
// 		if iv.Items[i].IsSelected {
// 			result[index] = iv.Items[i].Name
// 			index++
// 		}
// 	}

// 	return
// }

// func (iv *ItemView) getSelectedItemsPaths() (result []string) {
// 	result = make([]string, iv.getSelectedItemsCount())
// 	index := 0

// 	for i := 0; i < len(iv.Items); i++ {
// 		if iv.Items[i].IsSelected {
// 			result[index] = path.Join(iv.CurrentPath, iv.Items[i].Name)
// 			index++
// 		}
// 	}

// 	return
// }

// func (iv *ItemView) getSelectedItemsPathsAndNames() (result []string) {
// 	result = make([]string, iv.getSelectedItemsCount())
// 	index := 0

// 	for i := 0; i < len(iv.Items); i++ {
// 		if iv.Items[i].IsSelected {
// 			result[index] = path.Join(iv.CurrentPath, iv.Items[i].Name) + "|" + iv.Items[i].Name
// 			index++
// 		}
// 	}

// 	return
// }

func (iv *ItemView) favoriteIndex(fullPath string) int {
	for index, favorite := range iv.Favorites {
		if favorite.FullPath == fullPath {
			return index
		}
	}

	return -1
}

func (iv *ItemView) RemoveFavorite(fullPath string) {
	for i, favorite := range iv.Favorites {
		if favorite.FullPath == fullPath {
			iv.Favorites = append(iv.Favorites[:i], iv.Favorites[i+1:]...)
			break
		}
	}
}

func (iv *ItemView) CreateNewFile() {
	success, name := CreateNewFile(iv.CurrentPath)
	if !success {
		return
	}

	// // @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)
	iv.SetActiveByName(name)
	iv.RenameActive()
}

func (iv *ItemView) CreateNewFolder(updateView bool) string {
	success, name := CreateNewFolder(iv.CurrentPath, "New Folder")
	if !success {
		return ""
	}

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	if updateView {
		iv.ShowFolder(iv.CurrentPath)
	}

	iv.SetActiveByName(name)
	iv.RenameActive()

	return name
}

func (iv *ItemView) GroupSelectedFiles() {
	if !iv.SelectionMode && iv.getSelectedItemsCount() == 0 {
		return
	}

	newFolderName := iv.CreateNewFolder(false)

	for i := 0; i < len(iv.Items); i++ {
		if !iv.Items[i].IsSelected {
			continue
		}

		oldPath := path.Join(iv.CurrentPath, iv.Items[i].Name)
		newPath := path.Join(iv.CurrentPath, newFolderName, iv.Items[i].Name)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			NotifyError(err.Error())
			continue
		}
	}

	iv.SelectionMode = false

	iv.ShowFolder(iv.CurrentPath)
	iv.SetActiveByName(newFolderName)
}

func (iv *ItemView) Resize(rect sdl.Rect) {
	iv.Rect = rect
	iv.MaxItemsPerColumn = rect.H / iv.ItemHeight
}

func (iv *ItemView) Tick(input *Input) {
	if iv.Input.IsOpen {
		iv.Input.Tick(input)
		return
	}

	if iv.Mode == Mode_Goto {
		iv.handleInputGoto(input)
		return
	}

	if input.Escape {
		iv.SelectAll(false)
		iv.SelectionMode = false
		return
	}

	if input.Backspace {
		crumb := iv.App.Breadcrumbs.Pop()
		if crumb != "" {
			iv.GoOutside()
		}

		return
	}

	switch input.TypedCharacter {
	case 'h':
		iv.NavigateLeft()
	case 'j':
		iv.NavigateDown()
	case 'k':
		iv.NavigateUp()
	case 'l':
		iv.NavigateRight()
	case 'G':
		iv.NavigateLastInColumn()
	case 'g':
		if input.Ctrl {
			iv.GroupSelectedFiles()
		} else {
			iv.Mode = Mode_Goto
		}
	case '*':
		iv.MarkActiveAsFavorite()
	case 'x':
		if iv.getSelectedItemsCount() == 0 {
			iv.DeleteActive()
		} else {
			iv.DeleteSelected()
		}
	case 'y':
		if iv.getSelectedItemsCount() == 0 {
			iv.CopyActive(true)
		} else {
			// iv.CopySelected()
		}
	case 'p':
		iv.Paste()
	case 'D':
		iv.DuplicateActive()
	case 'r':
		iv.RenameActive()
	case 'v':
		iv.SelectActive()
	case 'V':
		iv.StartSelection()
	case 'a':
		if input.Ctrl {
			iv.SelectAll(true)
		}
	case 'i':
		iv.CreateNewFile()
	case 'I':
		iv.CreateNewFolder(true)
	case '`':
		iv.App.SelectFavorite(iv.favoritesToPaths())
	case '/':
		iv.App.FindInCurrentFolder(iv.itemsToNames())
	}
}

func (iv *ItemView) handleInputGoto(input *Input) {
	if input.Escape {
		iv.Mode = Mode_Normal
		return
	}

	if input.TypedCharacter == 0 {
		return
	}

	switch input.TypedCharacter {
	case 'd':
		iv.OpenItem(iv.Items[iv.ActiveItem].Name)
		iv.Mode = Mode_Normal
	case 'g':
		iv.App.ItemView.NavigateFirstInColumn()
		iv.Mode = Mode_Normal
	case 'h':
		iv.App.ShowFileInfo(iv.GetActiveFileInfo())
		iv.Mode = Mode_Normal
	default:
		iv.Mode = Mode_Normal
	}
}

func (iv *ItemView) Render(renderer *sdl.Renderer, app *App) {
	ivTheme := app.Theme.ItemViewTheme
	ifTheme := app.Theme.InputFieldTheme

	DrawRect3D(renderer, &iv.Rect, GetColor(ivTheme, "background_color"))

	var padding int32 = 10
	var itemPadding int32 = 5

	itemIndex := 0
	for i := 0; i < int(iv.Columns); i++ {
		for j := 0; j < int(iv.MaxItemsPerColumn); j++ {
			if itemIndex >= int(len(iv.Items)) {
				break
			}

			item := iv.Items[itemIndex]

			rect := sdl.Rect{
				X: iv.Rect.X + padding + int32(i)*iv.ItemWidth,
				Y: iv.Rect.Y + padding + iv.ItemHeight*int32(j),
				W: iv.ItemWidth,
				H: iv.ItemHeight,
			}

			font := app.Font

			if item.RenameInProgress {
				iv.Input.Render(renderer, rect, &font, ifTheme)
			} else {
				name := item.Name
				width := font.GetStringWidth(name)
				if width > iv.ItemWidth {
					name = font.ClipString(name, iv.ItemWidth-itemPadding*2)
					width = iv.ItemWidth - itemPadding*2
				}

				stringRect := sdl.Rect{
					X: rect.X + itemPadding,
					Y: rect.Y + (iv.ItemHeight-font.Size)/2,
					W: width,
					H: font.Size,
				}

				color := GetColor(ivTheme, "file_color")
				if item.Type == ItemTypeFolder {
					color = GetColor(ivTheme, "folder_color")
				} else if item.FileType == FileType_Exe {
					color = GetColor(ivTheme, "exe_color")
				} else if item.FileType == FileType_Image {
					color = GetColor(ivTheme, "image_color")
				}

				if itemIndex == int(iv.ActiveItem) {
					if HasColor(ivTheme, "active_background_color") {
						DrawRect(renderer, &rect, GetColor(ivTheme, "active_background_color"))
					}

					if HasColor(ivTheme, "active_background_border") {
						DrawRectOutline(renderer, &rect, GetColor(ivTheme, "active_background_border"))
					}

					color = GetColor(ivTheme, "active_file_color")
					if item.Type == ItemTypeFolder {
						color = GetColor(ivTheme, "active_folder_color")
					}
				} else if item.IsSelected {
					DrawRect(renderer, &rect, GetColor(ivTheme, "selected_background_color"))

					color = GetColor(ivTheme, "selected_file_color")
					if item.Type == ItemTypeFolder {
						color = GetColor(ivTheme, "selected_folder_color")
					}
				}

				DrawText(renderer, &font, name, &stringRect, color)

				if item.IsFavorite {
					iconRect := sdl.Rect{
						X: rect.X + rect.W - (itemPadding + iv.FavoriteIcon.Width),
						Y: rect.Y + (iv.ItemHeight-iv.FavoriteIcon.Height)/2,
						W: iv.FavoriteIcon.Width,
						H: iv.FavoriteIcon.Height,
					}

					DrawImage(renderer, iv.FavoriteIcon.Data, iconRect, GetColor(ivTheme, "favorite_icon_color"))
				}
			}

			itemIndex++
		}
	}
}
