package gee

import (
	"fmt"
	"net/http"
)

type router struct {
	handlers map[string]HandleFunc
}

func newRouter() *router {
	return &router{handlers: make(map[string]HandleFunc)}
}
func (e *router) addRoute(method, pattern string, handle HandleFunc) {
	key := method + "-" + pattern
	if _, ok := e.handlers[key]; !ok {
		e.handlers[key] = handle
	}
}

func (r *router) handle(ctx *Context) {
	key := ctx.Req.Method + "-" + ctx.Req.URL.Path
	if handle, ok := r.handlers[key]; ok {
		handle(ctx)
	} else {
		ctx.Writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(ctx.Writer, "404 not found %q \n", ctx.Req.URL)
	}
}
