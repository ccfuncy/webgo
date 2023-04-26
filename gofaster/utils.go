package gofaster

import (
	"fmt"
	"strings"
)

func SubStringLast(path string, sep string) string {
	index := strings.Index(path, sep)
	if index == -1 {
		return ""
	}
	fmt.Println(path[index+len(sep):])
	return path[index+len(sep):]
}
