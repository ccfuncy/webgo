package bytesconv

import "unsafe"

func StringToBytes(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{str, len(str)}))
}
