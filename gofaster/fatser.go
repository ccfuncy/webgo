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
}

func (group *routerGroup) handle(name string, method string, handleFunc HandlerFunc) {
	_, ok := group.handleMap[name]
	if !ok {
		group.handleMap[name] = make(map[string]HandlerFunc)
	}
	_, ok = group.handleMap[name][method]
	if ok {
		panic("相同路由下不能有相同请求")
	}
	group.handleMap[name][method] = handleFunc
	group.handleMapMethod[method] = append(group.handleMapMethod[method], name)
}

func (group *routerGroup) Any(name string, handleFunc HandlerFunc) {
	group.handle(name, ANY, handleFunc)
}
func (group *routerGroup) Get(name string, handleFunc HandlerFunc) {
	group.handle(name, http.MethodGet, handleFunc)
}
func (group *routerGroup) Post(name string, handleFunc HandlerFunc) {
	group.handle(name, http.MethodPost, handleFunc)
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
		for name, handlerFunc := range group.handleMap {
			if request.RequestURI == ("/" + group.name + name) {
				c := &Context{
					W: writer,
					R: request,
				}
				//支持任意请求
				handle, ok := handlerFunc[ANY]
				if ok {
					handle(c)
					return
				}

				handle, ok = handlerFunc[method]
				if ok {
					handle(c)
					return
				}
				writer.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(writer, "%s %s not allowed \n", request.RequestURI, method)
				return
			}

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
