package gee

import (
	"net/http"
)

func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

type HandleFunc func(ctx *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup //管理所有分组
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	context := NewContext(writer, request)
	e.router.handle(context)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
