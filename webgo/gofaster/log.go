package gofaster

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIp       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

type LoggerFormatter = func(params *LogFormatterParams) string

var DefaultWriter = os.Stdout
var defaultFormatter = func(params *LogFormatterParams) string {
	color := params.StatusCodeColor()
	resetColor := params.ResetColor()
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	if params.IsDisplayColor {
		return fmt.Sprintf("%s[gofaster]%s%s %v %s|%s %3d %s|%s %13v %s| %15s |%s %-7s %s %s %#v %s \n",
			yellow, resetColor,
			blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
			color, params.StatusCode, resetColor,
			red, params.Latency, resetColor,
			params.ClientIp,
			magenta, params.Method, resetColor,
			cyan, params.Path, resetColor)
	}
	return fmt.Sprintf("[gofaster] %v %3d | %13v | %15s | %-7s %#v \n",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		params.StatusCode,
		params.Latency,
		params.ClientIp,
		params.Method,
		params.Path)
}

// 中间件写法
func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}

func LoggingWithConfig(config LoggingConfig, next HandlerFunc) HandlerFunc {
	formatter := config.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}
	out := config.out
	if out == nil {
		out = DefaultWriter
	}
	return func(ctx *Context) {
		r := ctx.R
		//前置
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		next(ctx)
		//后置
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		clientIp := net.ParseIP(ip)
		method := r.Method
		statusCode := ctx.StatusCode
		if raw != "" {
			path = path + "?" + raw
		}
		param := &LogFormatterParams{
			Request:        r,
			TimeStamp:      stop,
			StatusCode:     statusCode,
			Latency:        latency,
			ClientIp:       clientIp,
			Method:         method,
			Path:           path,
			IsDisplayColor: out == os.Stdout,
		}
		fmt.Fprint(out, formatter(param))
	}
}
