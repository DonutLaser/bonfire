package main

import (
	"os"
	"path"
	"strings"
)

type Settings struct {
	Favorites []string
}

func NewSettings() Settings {
	dir, err := os.UserConfigDir()
	if err != nil {
		NotifyError(err.Error())

		return Settings{
			Favorites: []string{},
		}
	}

	fullPath := path.Join(dir, "bonfire", "settings.bfs")

	exists := DoesFileExist(fullPath)
	if exists {
		return loadSettings(fullPath)
	}

	result := Settings{
		Favorites: []string{},
	}

	result.Save(true)

	return result
}

func loadSettings(fullPath string) (result Settings) {
	data := ReadFile(fullPath)

	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, ":favorite") {
			result.Favorites = append(result.Favorites, line[10:])
		}
	}

	return
}

func (s *Settings) Save(createFolder bool) {
	dir, err := os.UserConfigDir()
	if err != nil {
		NotifyError(err.Error())
		return
	}

	var sb strings.Builder
	for _, favorite := range s.Favorites {
		sb.WriteString(":favorite ")
		sb.WriteString(favorite)
		sb.WriteByte('\n')
	}

	if createFolder {
		success, _ := CreateNewFolder(dir, "bonfire")
		if success {
			WriteFile(path.Join(dir, "bonfire", "settings.bfs"), sb.String())
		}
	} else {
		WriteFile(path.Join(dir, "bonfire", "settings.bfs"), sb.String())
	}
}

func (s *Settings) AddFavorite(fullPath string) {
	s.Favorites = append(s.Favorites, fullPath)
}

func (s *Settings) RemoveFavorite(fullPath string) {
	s.Favorites = Remove(s.Favorites, fullPath)
}
