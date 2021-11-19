package main

import (
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ItemType int32

const (
	Item_Type_File ItemType = iota
	Item_Type_Folder
)

type Clipboard struct {
	FullPath string
	Name     string
	Type     ItemType
}

type Item struct {
	Type ItemType
	Name string

	IsSelected       bool
	IsFavorite       bool
	RenameInProgress bool
}

type ItemView struct {
	Items        []Item
	ActiveItem   int32
	ActiveColumn int32
	CurrentPath  string

	Favorites []string
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
			break
		}
	}
}

func (iv *ItemView) ShowFolder(fullPath string) bool {
	dir, err := os.Open(fullPath)
	if err != nil {
		NotifyError(err.Error())
		return false
	}
	defer dir.Close()

	items, err := dir.Readdir(-1)
	if err != nil {
		NotifyError(err.Error())
		return false
	}

	var files []Item
	var folders []Item

	for _, file := range items {
		if file.IsDir() {
			folders = append(folders, Item{
				Type: Item_Type_Folder,
				Name: file.Name(),
			})
		} else {
			files = append(files, Item{
				Type: Item_Type_File,
				Name: file.Name(),
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
		p := path.Join(fullPath, item.Name)
		iv.Items[index].IsFavorite = IndexOf(iv.Favorites, p) >= 0
	}

	iv.Columns = int32(len(iv.Items)) / iv.MaxItemsPerColumn
	iv.Columns++ // Basically ceiling the number

	iv.ActiveItem = 0
	iv.CurrentPath = fullPath

	return true
}

func (iv *ItemView) OpenFolder(name string) {
	iv.SetActiveByName(name)

	activeFolder := iv.GoInside()
	if activeFolder != "" {
		iv.App.Breadcrumbs.Push(activeFolder)
	}
}

func (iv *ItemView) NavigateDown() {
	if iv.ActiveItem < int32(len(iv.Items)-1) {
		iv.ActiveItem++
		iv.updateSelection()
	}
}

func (iv *ItemView) NavigateUp() {
	if iv.ActiveItem > 0 {
		iv.ActiveItem--
		iv.updateSelection()
	}
}

func (iv *ItemView) NavigateRight() {
	if iv.ActiveColumn >= iv.Columns {
		return
	}

	iv.ActiveColumn++

	if iv.ActiveItem+iv.MaxItemsPerColumn >= int32(len(iv.Items)) {
		iv.ActiveColumn--
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
	if item.Type == Item_Type_Folder {
		result = item.Name
		iv.ShowFolder(path.Join(iv.CurrentPath, item.Name))
	}

	return
}

func (iv *ItemView) GoOutside() {
	split := strings.Split(iv.CurrentPath, "/")
	iv.CurrentPath = strings.Join(split[:len(split)-1], "/")

	iv.ShowFolder(iv.CurrentPath)
}

func (iv *ItemView) MarkActiveAsFavorite() {
	if iv.Items[iv.ActiveItem].IsFavorite {
		iv.Items[iv.ActiveItem].IsFavorite = false
		iv.Favorites = Remove(iv.Favorites, path.Join(iv.CurrentPath, iv.Items[iv.ActiveItem].Name))
	} else {
		iv.Favorites = append(iv.Favorites, path.Join(iv.CurrentPath, iv.Items[iv.ActiveItem].Name))
		iv.Items[iv.ActiveItem].IsFavorite = true
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
	iv.Clipboard.FullPath = path.Join(iv.CurrentPath, iv.Items[iv.ActiveItem].Name)
	iv.Clipboard.Type = iv.Items[iv.ActiveItem].Type

	if showNotification {
		NotifyInfo("Copied " + iv.Clipboard.FullPath)
	}
}

func (iv *ItemView) Paste() {
	if iv.Clipboard.Type == Item_Type_Folder {
		// @TODO (!important) implement pasting a folder
	} else if iv.Clipboard.Type == Item_Type_File {
		name := iv.getAvailableFileName(iv.Clipboard.Name)

		data, err := os.ReadFile(iv.Clipboard.FullPath)
		if err != nil {
			NotifyError(err.Error())
			return
		}

		err = os.WriteFile(path.Join(iv.CurrentPath, name), data, 0644)
		if err != nil {
			NotifyError(err.Error())
			return
		}

		iv.ShowFolder(iv.CurrentPath)
		iv.SetActiveByName(name)
	}
}

func (iv *ItemView) DuplicateActive() {
	if iv.Items[iv.ActiveItem].Type == Item_Type_Folder {

	} else if iv.Items[iv.ActiveItem].Type == Item_Type_File {
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

func (iv *ItemView) getAvailableFileName(baseName string) (result string) {
	// @TODO (!important) this should not add the numbers after the extension making the extension unusable

	fullPath := path.Join(iv.CurrentPath, baseName)

	fileDoesNotExist := true
	count := 0
	for fileDoesNotExist {
		_, err := os.Stat(fullPath)
		if err == nil {
			count += 1
			fullPath = path.Join(iv.CurrentPath, baseName+" ("+strconv.Itoa(count)+")")
		} else {
			fileDoesNotExist = false
			result = baseName + " (" + strconv.Itoa(count) + ")"
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

func (iv *ItemView) getSelectedItems() (result []string) {
	result = make([]string, iv.getSelectedItemsCount())
	index := 0

	for i := 0; i < len(iv.Items); i++ {
		if iv.Items[i].IsSelected {
			result[index] = iv.Items[i].Name
			index++
		}
	}

	return
}

func (iv *ItemView) getSelectedItemsPaths() (result []string) {
	result = make([]string, iv.getSelectedItemsCount())
	index := 0

	for i := 0; i < len(iv.Items); i++ {
		if iv.Items[i].IsSelected {
			result[index] = path.Join(iv.CurrentPath, iv.Items[i].Name)
			index++
		}
	}

	return
}

func (iv *ItemView) getSelectedItemsPathsAndNames() (result []string) {
	result = make([]string, iv.getSelectedItemsCount())
	index := 0

	for i := 0; i < len(iv.Items); i++ {
		if iv.Items[i].IsSelected {
			result[index] = path.Join(iv.CurrentPath, iv.Items[i].Name) + "|" + iv.Items[i].Name
			index++
		}
	}

	return
}

func (iv *ItemView) CreateNewFile() {
	name := iv.getAvailableFileName("New File")

	err := os.WriteFile(path.Join(iv.CurrentPath, name), []byte(""), 0644)
	if err != nil {
		NotifyError(err.Error())
		return
	}

	// // @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)
	iv.SetActiveByName(name)
	// // @TODO maybe automatically initiate rename of the new item
}

func (iv *ItemView) CreateNewFolder(updateView bool) string {
	name := iv.getAvailableFileName("New Folder")

	err := os.Mkdir(path.Join(iv.CurrentPath, name), 0755)
	if err != nil {
		NotifyError(err.Error())
		return ""
	}

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	if updateView {
		iv.ShowFolder(iv.CurrentPath)
	}

	iv.SetActiveByName(name)
	// @TODO maybe automatically initiate rename of the new item

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
		// @TODO (!important) this should retain the last position so that when you go back, the active item doesn't always become 0
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
			// @TODO (!important) show somehow that we are in the middle of changing drives
		}
	case 'm':
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
		iv.App.SelectFavorite(iv.Favorites)
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
		activeFolder := iv.GoInside()
		if activeFolder != "" {
			iv.App.Breadcrumbs.Push(activeFolder)
		}

		iv.Mode = Mode_Normal
	case 'g':
		iv.App.ItemView.NavigateFirstInColumn()
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
				if item.Type == Item_Type_Folder {
					color = GetColor(ivTheme, "folder_color")
				}

				if itemIndex == int(iv.ActiveItem) {
					DrawRect(renderer, &rect, GetColor(ivTheme, "active_background_color"))

					color = GetColor(ivTheme, "active_file_color")
					if item.Type == Item_Type_Folder {
						color = GetColor(ivTheme, "active_folder_color")
					}
				} else if item.IsSelected {
					DrawRect(renderer, &rect, GetColor(ivTheme, "selected_background_color"))

					color = GetColor(ivTheme, "selected_file_color")
					if item.Type == Item_Type_Folder {
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
