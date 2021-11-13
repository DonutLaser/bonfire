package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type QuickOpen struct {
	IsOpen      bool
	MaxItems    int32
	ItemsToShow int32 // Actual count of items we will show. Will never be more than MaxItems.
	Items       []string
	ActiveItem  int32
	Results     []string

	InputField *InputField
	OnSubmit   func(string)

	Rect          sdl.Rect
	BasePadding   int32
	InsidePadding int32
	ItemHeight    int32

	BackgroundColor           sdl.Color
	ItemColor                 sdl.Color
	ActiveItemBackgroundColor sdl.Color
}

func NewQuickOpen(rect sdl.Rect) *QuickOpen {
	return &QuickOpen{
		MaxItems:                  5,
		ItemsToShow:               0,
		InputField:                nil,
		OnSubmit:                  nil,
		Rect:                      rect,
		BasePadding:               8,
		InsidePadding:             5,
		ItemHeight:                24,
		BackgroundColor:           sdl.Color{R: 20, G: 27, B: 39, A: 255},
		ItemColor:                 sdl.Color{R: 216, G: 216, B: 216, A: 255},
		ActiveItemBackgroundColor: sdl.Color{R: 44, G: 50, B: 61, A: 255},
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
			}
		} else if input.TypedCharacter == 'k' {
			if q.ActiveItem > 0 {
				q.ActiveItem--
			}
		}

		return
	}

	if input.TypedCharacter == '\n' || input.TypedCharacter == '\t' {
		q.Submit()
		return
	}

	q.InputField.Tick(input)
}

func (q *QuickOpen) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, font *Font) {
	q.InputField.Render(renderer, parentRect.X+(parentRect.W-q.Rect.W)/2, parentRect.Y+100, font)

	baseRect := sdl.Rect{
		X: parentRect.X + (parentRect.W-q.Rect.W)/2,
		Y: parentRect.Y + 100 + q.InputField.Rect.H,
		W: q.Rect.W,
		H: q.BasePadding*2 + q.ItemsToShow*q.ItemHeight,
	}
	DrawRect3D(renderer, &baseRect, q.BackgroundColor)

	for i := 0; i < int(q.ItemsToShow); i++ {
		baseItemRect := sdl.Rect{
			X: baseRect.X + q.BasePadding,
			Y: baseRect.Y + q.BasePadding + int32(i)*q.ItemHeight,
			W: baseRect.W - q.BasePadding*2,
			H: q.ItemHeight,
		}

		if i == int(q.ActiveItem) {
			DrawRect(renderer, &baseItemRect, q.ActiveItemBackgroundColor)
		}

		value := q.Results[i]
		textWidth := font.GetStringWidth(value)
		textRect := sdl.Rect{
			X: baseItemRect.X + q.InsidePadding,
			Y: baseItemRect.Y + (baseItemRect.H-font.Size)/2,
			W: textWidth,
			H: font.Size,
		}
		DrawText(renderer, font, value, &textRect, q.ItemColor)
	}
}
