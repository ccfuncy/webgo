package gofaster

import (
	"strings"
	"unicode"
)

func SubStringLast(path string, sep string) string {
	index := strings.Index(path, sep)
	if index == -1 {
		return ""
	}
	//fmt.Println(path[index+len(sep):])
	return path[index+len(sep):]
}

func IsASCII(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
