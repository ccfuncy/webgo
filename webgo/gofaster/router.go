package gofaster

import "net/http"

const ANY = "Any"

type HandlerFunc func(ctx *Context)
type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc
type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		handleMap:       make(map[string]map[string]HandlerFunc),
		name:            name,
		handleMapMethod: make(map[string][]string),
		treeNode: &treeNode{
			name:      "/",
			childrens: make([]*treeNode, 0),
		},
	}
	r.routerGroups = append(r.routerGroups, group)
	return group
}

// 分组路由
type routerGroup struct {
	handleMap       map[string]map[string]HandlerFunc //路由对应的函数 map["name"]["get"] func
	handleMapMethod map[string][]string               //方法对应路由
	name            string
	treeNode        *treeNode //前缀树 动态路由 前缀树方式实现
	middlewares     []MiddlewareFunc
}

func (group *routerGroup) handle(name string, method string, handlerFunc HandlerFunc) {
	_, ok := group.handleMap[name]
	if !ok {
		group.handleMap[name] = make(map[string]HandlerFunc)
	}
	_, ok = group.handleMap[name][method]
	if ok {
		panic("相同路由下不能有相同请求")
	}
	group.handleMap[name][method] = handlerFunc
	group.handleMapMethod[method] = append(group.handleMapMethod[method], name)
	//前面是静态路由的请求方式 ，下面是动态路由的方式
	group.treeNode.Put(name)
}

func (group *routerGroup) Any(name string, handlerFunc HandlerFunc) {
	group.handle(name, ANY, handlerFunc)
}
func (group *routerGroup) Get(name string, handlerFunc HandlerFunc) {
	group.handle(name, http.MethodGet, handlerFunc)
}
func (group *routerGroup) Post(name string, handlerFunc HandlerFunc) {
	group.handle(name, http.MethodPost, handlerFunc)
}
func (group *routerGroup) Delete(name string, handlerFunc HandlerFunc) {
	group.handle(name, http.MethodDelete, handlerFunc)
}
func (group *routerGroup) Put(name string, handlerFunc HandlerFunc) {
	group.handle(name, http.MethodPut, handlerFunc)
}
func (group *routerGroup) Patch(name string, handlerFunc HandlerFunc) {
	group.handle(name, http.MethodPatch, handlerFunc)
}

func (group *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	group.middlewares = append(group.middlewares, middlewareFunc...)
}

func (group *routerGroup) methodHandle(h HandlerFunc, ctx *Context) {
	//中间件处理
	if group.middlewares != nil {
		for _, middleware := range group.middlewares {
			h = middleware(h)
		}
	}
	h(ctx)
}
