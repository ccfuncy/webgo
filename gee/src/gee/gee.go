package gee

import (
	"net/http"
)

func New() *Engine {
	return &Engine{router: newRouter()}
}

type HandleFunc func(ctx *Context)

type Engine struct {
	router *router
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	context := NewContext(writer, request)
	e.router.handle(context)
}

func (e *Engine) Post(pattern string, handle HandleFunc) {
	e.router.addRoute("POST", pattern, handle)
}
func (e *Engine) Get(pattern string, handle HandleFunc) {
	e.router.addRoute("GET", pattern, handle)
}
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
