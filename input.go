package main

type Input struct {
	TypedCharacter byte
	Backspace      bool
	Escape         bool
	Ctrl           bool
	Alt            bool
	Shift          bool
	F11            bool
}

func (input *Input) Clear() {
	input.TypedCharacter = 0
	input.Backspace = false
	input.Escape = false
	input.F11 = false
}

type Shortcut struct {
	Ctrl     bool
	Alt      bool
	Callback func()
}
