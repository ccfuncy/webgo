package log

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (l *TextFormatter) Format(parameter *LoggerParameter) string {
	var fields strings.Builder
	if parameter.LoggerFields != nil {
		var count = len(parameter.LoggerFields)
		var i = 1
		for k, v := range parameter.LoggerFields {
			fmt.Fprintf(&fields, "%s=%v", k, v)
			if count-1 == i {
				fmt.Fprint(&fields, ",")
				i++
			}
		}
	}
	if parameter.IsColor {
		msgColor := l.MsgColor(parameter.level)
		levelColor := l.LevelColor(parameter.level)
		return fmt.Sprintf("[gofaster] %v | level=%s %s %s| msg=%s %#v %s %s \n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			levelColor, parameter.level.Level(), reset,
			msgColor, parameter.Msg, reset, fields.String())
	}
	return fmt.Sprintf("[gofaster] %v | level=%s | msg=%#v %s \n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		parameter.level.Level(),
		parameter.Msg,
		fields.String(),
	)
}
func (l *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}
func (l *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
	case LevelError:
		return red
	default:
		return cyan
	}
}
