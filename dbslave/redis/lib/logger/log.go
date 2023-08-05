package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	reset     = "\033[0m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
)

type LoggerLevel int

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelError:
		return "ERROR"
	case LevelInfo:
		return "INFO"
	default:
		return ""
	}
}

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type Fields map[string]any
type LoggerWriter struct {
	out   io.Writer
	level LoggerLevel
}
type Logger struct {
	Formatter    LoggerFormatter
	Outs         []*LoggerWriter
	LoggerFields Fields
	LogPath      string
	LogFileSize  int64
}

func (l *Logger) SetPath(logPath string) {
	l.LogPath = logPath
	l.Outs = append(l.Outs, &LoggerWriter{
		out:   FileWriter(path.Join(logPath, "all.log")),
		level: -1,
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		out:   FileWriter(path.Join(logPath, "info.log")),
		level: LevelInfo,
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		out:   FileWriter(path.Join(logPath, "error.log")),
		level: LevelError,
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		out:   FileWriter(path.Join(logPath, "debug.log")),
		level: LevelDebug,
	})
}

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	writer := &LoggerWriter{
		out:   os.Stdout,
		level: LevelDebug,
	}
	logger.Outs = append(logger.Outs, writer)
	logger.Formatter = &TextFormatter{}
	return logger
}

func (l *Logger) Print(level LoggerLevel, msg any) {
	for _, out := range l.Outs {
		parameter := LoggerParameter{
			level:        level,
			IsColor:      out.out == os.Stdout,
			LoggerFields: l.LoggerFields,
			Msg:          msg,
		}
		if os.Stdout == out.out {
			fmt.Fprint(out.out, l.Formatter.Format(&parameter))
		} else {
			if out.level == -1 || out.level == level {
				fmt.Fprint(out.out, l.Formatter.Format(&parameter))
			}
			l.checkOutSize(out)
		}
	}
}

func (l *Logger) checkOutSize(out *LoggerWriter) {
	file := out.out.(*os.File)
	if file != nil {
		stat, err := file.Stat()
		if err != nil {
			log.Println(err)
			return
		}
		size := stat.Size()
		if l.LogFileSize < 0 {
			l.LogFileSize = 100 << 20 //100M
		}
		if size >= l.LogFileSize {
			_, name := path.Split(stat.Name())
			fileName := name[:strings.Index(name, ".")]
			out.out = FileWriter(path.Join(l.LogPath, JoinStrings(fileName, ".", time.Now().UnixMilli(), ".log")))
		}
	}
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}
func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}
func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}
func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		LoggerFields: fields,
		LogPath:      l.LogPath,
		LogFileSize:  l.LogFileSize,
	}
}

type LoggerParameter struct {
	level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
	Msg          any
}

type LoggerFormatter interface {
	Format(*LoggerParameter) string
	MsgColor(LoggerLevel) string
	LevelColor(LoggerLevel) string
}

func FileWriter(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return file
}
