package gee

import (
	"fmt"
	"net/http"
	"strings"
)

type router struct {
	root     map[string]*node
	handlers map[string]HandleFunc
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandleFunc),
		root:     make(map[string]*node),
	}
}

func parsePattern(pattern string) []string {
	res := make([]string, 0)
	splits := strings.Split(pattern, "/")

	for _, split := range splits {
		if split != "" {
			res = append(res, split)
			if split[0] == '*' {
				break
			}
		}
	}
	return res
}

func (r *router) addRoute(method, pattern string, handle HandleFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	if _, ok := r.root[method]; !ok {
		r.root[method] = &node{}
	}
	r.root[method].insert(pattern, parts, 0)
	r.handlers[key] = handle
}

func (r *router) getRoute(method, pattern string) (*node, map[string]string) {
	param := make(map[string]string)
	searchParts := parsePattern(pattern)
	if _, ok := r.root[method]; !ok {
		return nil, nil
	}
	root := r.root[method]
	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				param[part[1:]] = searchParts[index]
			}
			if part[0] == '*' {
				param[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, param
	}
	return nil, nil
}

func (r *router) handle(ctx *Context) {
	n, param := r.getRoute(ctx.Method, ctx.Path)
	if n != nil {
		key := ctx.Req.Method + "-" + n.pattern
		ctx.Params = param

		r.handlers[key](ctx)
	} else {
		ctx.Writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(ctx.Writer, "404 not found %q \n", ctx.Req.URL)
	}
}
