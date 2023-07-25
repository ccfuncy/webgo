package orm

import (
	"reflect"
	"strings"
)

func IsAutoId(id any) bool {
	value := reflect.ValueOf(id)
	switch value.Kind() {
	case reflect.Int32:
		id := id.(int32)
		if id <= 0 {
			return true
		}
	case reflect.Int64:
		id := id.(int64)
		if id <= 0 {
			return true
		}
	case reflect.Int:
		id := id.(int)
		if id <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

func Name(name string) string {
	//UserName->User_Name
	var names = name[:]
	var lastIndex = 0
	var sb strings.Builder
	for index, value := range names {
		if value >= 65 && value <= 90 {
			//大写字母
			if index == 0 {
				continue
			}
			sb.WriteString(name[lastIndex:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	sb.WriteString(name[lastIndex:])
	return sb.String()
}
