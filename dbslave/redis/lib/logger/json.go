package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
}

func (j JsonFormatter) Format(parameter *LoggerParameter) string {
	if parameter.LoggerFields == nil {
		parameter.LoggerFields = make(Fields)
	}
	parameter.LoggerFields["log_time"] = time.Now().Format("2006/01/02 - 15:04:05")
	parameter.LoggerFields["msg"] = parameter.Msg
	parameter.LoggerFields["log_level"] = parameter.level.Level()
	marshal, err := json.Marshal(parameter.LoggerFields)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s\n", marshal)
}

func (j JsonFormatter) MsgColor(LoggerLevel) string {
	return cyan
}

func (j JsonFormatter) LevelColor(LoggerLevel) string {
	return cyan
}
