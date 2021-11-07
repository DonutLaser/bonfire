package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func getCharacter(shift bool, lowercase byte, uppercase byte) byte {
	if shift {
		return uppercase
	}

	return lowercase
}

func keyToCharacter(key sdl.Keycode, mod uint16) byte {
	shift := ((mod & sdl.KMOD_LSHIFT) | (mod & sdl.KMOD_RSHIFT) | mod&sdl.KMOD_CAPS) != 0

	switch key {
	case sdl.K_q:
		return getCharacter(shift, 'q', 'Q')
	case sdl.K_w:
		return getCharacter(shift, 'w', 'W')
	case sdl.K_e:
		return getCharacter(shift, 'e', 'E')
	case sdl.K_r:
		return getCharacter(shift, 'r', 'R')
	case sdl.K_t:
		return getCharacter(shift, 't', 'T')
	case sdl.K_y:
		return getCharacter(shift, 'y', 'Y')
	case sdl.K_u:
		return getCharacter(shift, 'u', 'U')
	case sdl.K_i:
		return getCharacter(shift, 'i', 'I')
	case sdl.K_o:
		return getCharacter(shift, 'o', 'O')
	case sdl.K_p:
		return getCharacter(shift, 'p', 'P')
	case sdl.K_LEFTBRACKET:
		return getCharacter(shift, '[', '{')
	case sdl.K_RIGHTBRACKET:
		return getCharacter(shift, ']', '}')
	case sdl.K_BACKSLASH:
		return getCharacter(shift, '\\', '|')
	case sdl.K_a:
		return getCharacter(shift, 'a', 'A')
	case sdl.K_s:
		return getCharacter(shift, 's', 'S')
	case sdl.K_d:
		return getCharacter(shift, 'd', 'D')
	case sdl.K_f:
		return getCharacter(shift, 'f', 'F')
	case sdl.K_g:
		return getCharacter(shift, 'g', 'G')
	case sdl.K_h:
		return getCharacter(shift, 'h', 'H')
	case sdl.K_j:
		return getCharacter(shift, 'j', 'J')
	case sdl.K_k:
		return getCharacter(shift, 'k', 'K')
	case sdl.K_l:
		return getCharacter(shift, 'l', 'L')
	case sdl.K_SEMICOLON:
		return getCharacter(shift, ';', ':')
	case sdl.K_QUOTE:
		return getCharacter(shift, '\'', '"')
	case sdl.K_z:
		return getCharacter(shift, 'z', 'Z')
	case sdl.K_x:
		return getCharacter(shift, 'x', 'X')
	case sdl.K_c:
		return getCharacter(shift, 'c', 'C')
	case sdl.K_v:
		return getCharacter(shift, 'v', 'V')
	case sdl.K_b:
		return getCharacter(shift, 'b', 'B')
	case sdl.K_n:
		return getCharacter(shift, 'n', 'N')
	case sdl.K_m:
		return getCharacter(shift, 'm', 'M')
	case sdl.K_COMMA:
		return getCharacter(shift, ',', '<')
	case sdl.K_PERIOD:
		return getCharacter(shift, '.', '>')
	case sdl.K_SLASH:
		return getCharacter(shift, '/', '?')
	case sdl.K_SPACE:
		return getCharacter(shift, ' ', ' ')
	case sdl.K_BACKQUOTE:
		return getCharacter(shift, '`', '~')
	case sdl.K_1:
		return getCharacter(shift, '1', '!')
	case sdl.K_2:
		return getCharacter(shift, '2', '@')
	case sdl.K_3:
		return getCharacter(shift, '3', '#')
	case sdl.K_4:
		return getCharacter(shift, '4', '$')
	case sdl.K_5:
		return getCharacter(shift, '5', '%')
	case sdl.K_6:
		return getCharacter(shift, '6', '^')
	case sdl.K_7:
		return getCharacter(shift, '7', '&')
	case sdl.K_8:
		return getCharacter(shift, '8', '*')
	case sdl.K_9:
		return getCharacter(shift, '9', '(')
	case sdl.K_0:
		return getCharacter(shift, '0', ')')
	case sdl.K_MINUS:
		return getCharacter(shift, '-', '_')
	case sdl.K_EQUALS:
		return getCharacter(shift, '=', '+')
	case sdl.K_RETURN:
		return getCharacter(shift, '\n', '\n')
	case sdl.K_TAB:
		return getCharacter(shift, '\t', '\t')
	}

	return 0
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkError(err)
	defer sdl.Quit()

	err = ttf.Init()
	checkError(err)
	defer ttf.Quit()

	window, err := sdl.CreateWindow("bonfire", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_RESIZABLE)
	checkError(err)
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	checkError(err)
	defer renderer.Destroy()

	windowWidth, windowHeight := window.GetSize()

	app := NewApp(windowWidth, windowHeight)
	input := Input{}

	running := true
	for running {
		input.Clear()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				keycode := t.Keysym.Sym

				switch keycode {
				case sdl.K_LCTRL:
					fallthrough
				case sdl.K_RCTRL:
					input.Ctrl = t.Type == sdl.KEYDOWN
				case sdl.K_LALT:
					fallthrough
				case sdl.K_RALT:
					input.Alt = t.Type == sdl.KEYDOWN
				case sdl.K_LSHIFT:
					fallthrough
				case sdl.K_RSHIFT:
					input.Shift = t.Type == sdl.KEYDOWN
				case sdl.K_BACKSPACE:
					if t.State != sdl.RELEASED {
						input.Backspace = true
					}
				case sdl.K_ESCAPE:
					if t.State != sdl.RELEASED {
						input.Escape = true
					}
				default:
					if t.State != sdl.RELEASED {
						input.TypedCharacter = keyToCharacter(keycode, t.Keysym.Mod)
					}
				}
			}
		}

		app.Tick(&input)
		app.Render(renderer)
	}
}
