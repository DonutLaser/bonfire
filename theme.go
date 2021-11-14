package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// @TODO (!important) support custom icons

type Subtheme map[string]interface{}

type Theme struct {
	BreadcrumbsTheme  Subtheme
	ItemViewTheme     Subtheme
	InputFieldTheme   Subtheme
	QuickOpenTheme    Subtheme
	NotificationTheme Subtheme
}

func LoadTheme(themeName string) (result *Theme) {
	data := ReadFile(fmt.Sprintf("./assets/themes/%s.bft", themeName))

	result = &Theme{
		BreadcrumbsTheme:  Subtheme{},
		ItemViewTheme:     Subtheme{},
		InputFieldTheme:   Subtheme{},
		QuickOpenTheme:    Subtheme{},
		NotificationTheme: Subtheme{},
	}
	currentSubtheme := result.BreadcrumbsTheme

	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line[0] == '@' {
			if strings.Contains(line, "Breadcrumbs") {
				currentSubtheme = result.BreadcrumbsTheme
			} else if strings.Contains(line, "ItemView") {
				currentSubtheme = result.ItemViewTheme
			} else if strings.Contains(line, "InputField") {
				currentSubtheme = result.InputFieldTheme
			} else if strings.Contains(line, "QuickOpen") {
				currentSubtheme = result.QuickOpenTheme
			} else if strings.Contains(line, "Notification") {
				currentSubtheme = result.NotificationTheme
			}
		} else {
			key, value := getKeyValue(line)
			currentSubtheme[key] = getColor(value)
		}
	}

	return
}

func GetColor(subtheme Subtheme, key string) sdl.Color {
	if value, ok := subtheme[key]; ok {
		return value.(sdl.Color)
	}

	return sdl.Color{R: 255, G: 0, B: 255, A: 255}
}

func getKeyValue(text string) (key string, value string) {
	split := strings.Split(text, " = ")
	key = split[0]
	value = split[1]

	return
}

func getColor(text string) sdl.Color {
	cc := strings.Split(text, " ") // Color components
	r, _ := strconv.Atoi(cc[0])
	g, _ := strconv.Atoi(cc[1])
	b, _ := strconv.Atoi(cc[2])

	return sdl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
