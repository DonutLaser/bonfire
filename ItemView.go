package main

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ItemType int32

const (
	Item_Type_File ItemType = iota
	Item_Type_Folder
)

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
	Favorites    []string

	Rect              sdl.Rect
	ItemHeight        int32
	ItemWidth         int32
	MaxItemsPerColumn int32
	Columns           int32

	FavoriteIcon   *Image
	Input          *InlineInputField
	ConsumingInput bool
	SelectionMode  bool
	SelectionStart int32
}

func NewItemView(rect sdl.Rect, favoriteIcon *Image) *ItemView {
	return &ItemView{
		ActiveItem:        -1,
		ActiveColumn:      0,
		Rect:              rect,
		ItemHeight:        24,
		ItemWidth:         394,
		MaxItemsPerColumn: rect.H / 24,
		FavoriteIcon:      favoriteIcon,
		Input:             NewInlineInputField(),
	}
}

func (iv *ItemView) Close() {
	iv.FavoriteIcon.Unload()
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
	checkError(err)

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)

	// @TODO (!important) do not set the first item in the list as active. The item above the deleted one should become active instead
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
		checkError(err)
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

func (iv *ItemView) CreateNewFile() {
	// @TODO (!important) make sure file "New File" does not yet exist. If it does, add a number to the end of the name
	err := os.WriteFile(path.Join(iv.CurrentPath, "New File"), []byte(""), 0644)
	checkError(err)

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)

	// @TODO (!important) make the new item active
	// @TODO maybe automatically initiate rename of the new item
}

func (iv *ItemView) CreateNewFolder() {
	err := os.Mkdir(path.Join(iv.CurrentPath, "New Folder"), 0755)
	checkError(err)

	// @TODO (!important) not really efficient, better way would probably be to modify the existing items list instead of overriding it
	iv.ShowFolder(iv.CurrentPath)

	// @TODO (!important) make the new item active
	// @TODO maybe automatically initiate rename of the new item
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

	if input.Escape {
		iv.SelectAll(false)
		iv.SelectionMode = false
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
	case 'm':
		iv.MarkActiveAsFavorite()
	case 'x':
		iv.DeleteActive()
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
		iv.CreateNewFolder()
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
				continue
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
