package gee

import (
	"fmt"
	"net/http"
)

func New() *Engine {
	return &Engine{router: make(map[string]HandleFunc)}
}

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type Engine struct {
	router map[string]HandleFunc
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	key := request.Method + "-" + request.URL.Path
	if handle, ok := e.router[key]; ok {
		handle(writer, request)
	} else {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, "404 not found %q \n", request.URL)
	}
}

func (e *Engine) addRoute(method, pattern string, handle HandleFunc) {
	key := method + "-" + pattern
	if _, ok := e.router[key]; !ok {
		e.router[key] = handle
	}
}
func (e *Engine) Post(pattern string, handle HandleFunc) {
	e.addRoute("POST", pattern, handle)
}
func (e *Engine) Get(pattern string, handle HandleFunc) {
	e.addRoute("GET", pattern, handle)
}
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
