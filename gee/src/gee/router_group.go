package gee

import "log"

type RouterGroup struct {
	prefix     string
	middleware []HandleFunc
	pattern    *RouterGroup
	engine     *Engine
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	newGroup := &RouterGroup{
		prefix:  g.engine.prefix + prefix,
		pattern: g,
		engine:  g.engine,
	}
	g.engine.groups = append(g.engine.groups, newGroup)
	return newGroup
}

func (g *RouterGroup) addRoute(method, comp string, handle HandleFunc) {
	pattern := g.prefix + comp
	log.Printf("Route %4s - %s ", method, pattern)
	g.engine.router.addRoute(method, pattern, handle)
}
func (g *RouterGroup) Get(pattern string, handle HandleFunc) {
	g.addRoute("GET", pattern, handle)
}
func (g *RouterGroup) Post(pattern string, handle HandleFunc) {
	g.addRoute("GET", pattern, handle)
}

func (g *RouterGroup) Use(middleware ...HandleFunc) {
	g.middleware = append(g.middleware, middleware...)
}
