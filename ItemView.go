package main

import (
	"os"
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
}

type ItemView struct {
	Items        []Item
	ActiveItem   int32
	ActiveColumn int32
	CurrentPath  string

	Rect              sdl.Rect
	ItemHeight        int32
	ItemWidth         int32
	MaxItemsPerColumn int32
	Columns           int32

	BackgroundColor sdl.Color
	FolderColor     sdl.Color
	FileColor       sdl.Color
	ActiveItemColor sdl.Color
}

func NewItemView(rect sdl.Rect) *ItemView {
	return &ItemView{
		ActiveItem:        -1,
		ActiveColumn:      0,
		Rect:              rect,
		ItemHeight:        24,
		ItemWidth:         394,
		MaxItemsPerColumn: rect.H / 24,
		BackgroundColor:   sdl.Color{R: 20, G: 27, B: 39, A: 255},
		FolderColor:       sdl.Color{R: 254, G: 203, B: 0, A: 255},
		FileColor:         sdl.Color{R: 216, G: 216, B: 216, A: 255},
		ActiveItemColor:   sdl.Color{R: 44, G: 50, B: 61, A: 255},
	}
}

func (iv *ItemView) ShowFolder(path string) {
	dir, err := os.Open(path)
	checkError(err)
	defer dir.Close()

	items, err := dir.Readdir(-1)
	checkError(err)

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

	iv.Columns = int32(len(iv.Items)) / iv.MaxItemsPerColumn
	iv.Columns++ // Basically ceiling the number

	iv.ActiveItem = 0
	iv.CurrentPath = path
}

func (iv *ItemView) NavigateDown() {
	if iv.ActiveItem < int32(len(iv.Items)-1) {
		iv.ActiveItem++
	}
}

func (iv *ItemView) NavigateUp() {
	if iv.ActiveItem > 0 {
		iv.ActiveItem--
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
}

func (iv *ItemView) NavigateLeft() {
	if iv.ActiveColumn <= 0 {
		return
	}

	iv.ActiveColumn--
	iv.ActiveItem -= iv.MaxItemsPerColumn
}

func (iv *ItemView) NavigateFirstInColumn() {
	iv.ActiveItem = iv.ActiveColumn * iv.MaxItemsPerColumn
}

func (iv *ItemView) NavigateLastInColumn() {
	iv.ActiveItem = iv.ActiveColumn*iv.MaxItemsPerColumn + iv.MaxItemsPerColumn - 1
	if iv.ActiveItem >= int32(len(iv.Items)) {
		iv.ActiveItem = int32(len(iv.Items) - 1)
	}
}

func (iv *ItemView) GoInside() (result string) {
	if iv.ActiveItem < 0 || iv.ActiveItem >= int32(len(iv.Items)) {
		return
	}

	item := iv.Items[iv.ActiveItem]
	if item.Type == Item_Type_Folder {
		// @TODO (!important) write custom path functions
		var sb strings.Builder
		sb.WriteString(iv.CurrentPath)
		sb.WriteString("/")
		sb.WriteString(item.Name)

		result = item.Name
		iv.ShowFolder(sb.String())
	}

	return
}

func (iv *ItemView) GoOutside() {
	split := strings.Split(iv.CurrentPath, "/")
	iv.CurrentPath = strings.Join(split[:len(split)-1], "/")

	iv.ShowFolder(iv.CurrentPath)
}

func (iv *ItemView) Resize(rect sdl.Rect) {
	iv.Rect = rect
	iv.MaxItemsPerColumn = rect.H / iv.ItemHeight
}

func (iv *ItemView) Render(renderer *sdl.Renderer, font *Font) {
	DrawRect3D(renderer, &iv.Rect, iv.BackgroundColor)

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

			color := iv.FileColor
			if item.Type == Item_Type_Folder {
				color = iv.FolderColor
			}

			if itemIndex == int(iv.ActiveItem) {
				DrawRect(renderer, &rect, iv.ActiveItemColor)
			}
			DrawText(renderer, font, name, &stringRect, color)

			itemIndex++
		}

	}
}
