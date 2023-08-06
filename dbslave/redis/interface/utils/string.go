package utils

import (
	"redis/interface/database"
	"unsafe"
)

func StringToBytes(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{str, len(str)}))
}

func BytesToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func ToCmdLine(cmd ...string) database.Cmdline {
	args := make(database.Cmdline, len(cmd))
	for i, s := range cmd {
		args[i] = StringToBytes(s)
	}
	return args
}

func ToCmdLine2(command string, args ...[]byte) database.Cmdline {
	cmdline := make(database.Cmdline, len(args)+1)
	cmdline[0] = StringToBytes(command)
	for i, arg := range args {
		cmdline[i+1] = arg
	}
	return cmdline
}
