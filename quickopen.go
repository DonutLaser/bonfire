package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type QuickOpen struct {
	IsOpen            bool
	MaxItems          int32
	ItemsToShow       int32 // Actual count of items we will show. Will never be more than MaxItems.
	Items             []string
	ActiveItem        int32
	Results           []string
	ActiveItemChanged bool

	InputField *InputField
	OnSubmit   func(string)

	Rect          sdl.Rect
	BasePadding   int32
	InsidePadding int32
	ItemHeight    int32
}

func NewQuickOpen(rect sdl.Rect) *QuickOpen {
	return &QuickOpen{
		MaxItems:      5,
		ItemsToShow:   0,
		InputField:    nil,
		OnSubmit:      nil,
		Rect:          rect,
		BasePadding:   8,
		InsidePadding: 5,
		ItemHeight:    24,
	}
}

func (q *QuickOpen) Open(items []string, submitCallback func(string)) {
	if q.InputField == nil {
		q.InputField = NewInputField(sdl.Rect{X: 0, Y: 0, W: q.Rect.W, H: 40}, q.OnInput)
	}

	q.IsOpen = true
	q.Items = items
	q.Results = items

	q.ItemsToShow = int32(len(items))
	if q.ItemsToShow > q.MaxItems {
		q.ItemsToShow = q.MaxItems
	}

	q.OnSubmit = submitCallback
}

func (q *QuickOpen) Close() {
	q.InputField.Clear()
	q.OnSubmit = nil
	q.IsOpen = false
}

func (q *QuickOpen) Submit() {
	// @TODO (!important) handle files, not only folders
	if q.OnSubmit != nil {
		q.OnSubmit(q.Results[q.ActiveItem])
	}

	q.Close()
}

func (q *QuickOpen) OnInput(value string) {
	results := []string{}

	lastCount := q.ItemsToShow

	for _, item := range q.Items {
		if strings.Contains(item, value) {
			results = append(results, item)
		}
	}

	q.Results = results

	if len(q.Results) == 0 {
		q.ItemsToShow = int32(len(q.Items))
		q.Results = q.Items
	} else {
		q.ItemsToShow = int32(len(q.Results))
	}

	if q.ItemsToShow > q.MaxItems {
		q.ItemsToShow = q.MaxItems
	}

	if lastCount != q.ItemsToShow {
		q.ActiveItem = 0
	}
}

func (q *QuickOpen) Tick(input *Input) {
	if input.Escape {
		q.Close()
		return
	}

	if input.Alt {
		if input.TypedCharacter == 'j' {
			if q.ActiveItem < q.ItemsToShow-1 {
				q.ActiveItem++
				q.ActiveItemChanged = true
			}
		} else if input.TypedCharacter == 'k' {
			if q.ActiveItem > 0 {
				q.ActiveItem--
				q.ActiveItemChanged = true
			}
		}

		return
	} else if q.ActiveItemChanged {
		q.Submit()
		return
	}

	if input.TypedCharacter == '\n' || input.TypedCharacter == '\t' {
		q.Submit()
		return
	}

	q.InputField.Tick(input)
}

func (q *QuickOpen) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, font *Font, theme Subtheme, inputFieldTheme Subtheme) {
	q.InputField.Render(renderer, parentRect.X+(parentRect.W-q.Rect.W)/2, parentRect.Y+100, font, inputFieldTheme)

	if q.ItemsToShow == 0 {
		return
	}

	baseRect := sdl.Rect{
		X: parentRect.X + (parentRect.W-q.Rect.W)/2,
		Y: parentRect.Y + 100 + q.InputField.Rect.H,
		W: q.Rect.W,
		H: q.BasePadding*2 + q.ItemsToShow*q.ItemHeight,
	}
	DrawRect3D(renderer, &baseRect, GetColor(theme, "background_color"))

	for i := 0; i < int(q.ItemsToShow); i++ {
		baseItemRect := sdl.Rect{
			X: baseRect.X + q.BasePadding,
			Y: baseRect.Y + q.BasePadding + int32(i)*q.ItemHeight,
			W: baseRect.W - q.BasePadding*2,
			H: q.ItemHeight,
		}

		textColor := GetColor(theme, "item_color")
		if i == int(q.ActiveItem) {
			DrawRect(renderer, &baseItemRect, GetColor(theme, "active_item_background_color"))
			textColor = GetColor(theme, "active_item_color")
		}

		value := q.Results[i]
		textWidth := font.GetStringWidth(value)
		textRect := sdl.Rect{
			X: baseItemRect.X + q.InsidePadding,
			Y: baseItemRect.Y + (baseItemRect.H-font.Size)/2,
			W: textWidth,
			H: font.Size,
		}
		DrawText(renderer, font, value, &textRect, textColor)
	}
}
