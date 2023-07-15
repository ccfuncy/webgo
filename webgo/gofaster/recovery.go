package gofaster

import (
	"errors"
	"fmt"
	"gofaster/fserror"
	"net/http"
	"runtime"
	"strings"
)

func detailMsg(err any) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v\n", err))
	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

func Recovery(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				err2 := err.(error)
				if err2 != nil {
					var fsError *fserror.FsError
					if errors.As(err2, &fsError) {
						fsError.ExecResult()
						return
					}
				}
				ctx.E.Logger.Error(detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error!")
			}
		}()

		next(ctx)
	}
}
