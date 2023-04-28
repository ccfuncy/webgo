package gofaster

import (
	"fmt"
	"log"
	"net/http"
)

const ANY = "Any"

// 分组路由
type routerGroup struct {
	handleMap       map[string]map[string]HandlerFunc //路由对应的函数 map["name"]["get"] func
	handleMapMethod map[string][]string               //方法对应路由
	name            string
	treeNode        *treeNode //前缀树 动态路由 前缀树方式实现
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

type HandlerFunc func(ctx *Context)

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

type Engine struct {
	router
}

// 实现该接口，将所有请求转交给他处理分发
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	method := request.Method
	//先遍历分组路由
	for _, group := range e.routerGroups {
		///再遍历分组路由下的路由集合
		routerName := SubStringLast(request.RequestURI, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil && len(node.childrens) == 0 {
			c := &Context{
				W: writer,
				R: request,
			}
			//支持任意请求
			handle, ok := group.handleMap[node.routerName][ANY]
			if ok {
				handle(c)
				return
			}

			handle, ok = group.handleMap[node.routerName][method]
			if ok {
				handle(c)
				return
			}
			writer.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(writer, "%s %s not allowed \n", request.RequestURI, method)
			return
		}
	}
	writer.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(writer, "%s %s not found \n", request.RequestURI, method)
	return
}

func New() *Engine {
	return &Engine{router: router{}}
}

func (e *Engine) Run() {
	//for _, group := range e.routerGroups {
	//	for s, handlerFunc := range group.handleMap {
	//		http.HandleFunc("/"+group.name+s, handlerFunc)
	//	}
	//}
	http.Handle("/", e)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
