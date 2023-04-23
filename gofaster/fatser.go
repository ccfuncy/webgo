package gofaster

import (
	"log"
	"net/http"
)

type routerGroup struct {
	handleMap map[string]HandlerFunc
	name      string
}

func (group *routerGroup) Add(name string, handleFunc HandlerFunc) {
	group.handleMap[name] = handleFunc
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{handleMap: make(map[string]HandlerFunc),
		name: name}
	r.routerGroups = append(r.routerGroups, group)
	return group
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{router: router{}}
}

func (e *Engine) Run() {
	for _, group := range e.routerGroups {
		for s, handlerFunc := range group.handleMap {
			http.HandleFunc("/"+group.name+s, handlerFunc)
		}
	}

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
