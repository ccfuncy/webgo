package database

import "strings"

var cmdTable = make(map[string]*command)

type command struct {
	executor ExecFunc
	arity    int //参数数量
}

func RegisterCommand(name string, executor ExecFunc, arity int) {
	upper := strings.ToUpper(name)
	cmdTable[upper] = &command{
		executor: executor,
		arity:    arity,
	}
}
