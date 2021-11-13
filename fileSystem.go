package main

import (
	"os"
)

func ReadFile(fullPath string) string {
	contents, err := os.ReadFile(fullPath)
	checkError(err)

	return string(contents)
}
