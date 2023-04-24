package gofaster

import (
	"fmt"
	"log"
	"net/http"
)

const ANY = "Any"

// 分组路由
type routerGroup struct {
	handleMap       map[string]HandlerFunc //路由对应的函数
	handleMapMethod map[string][]string    //方法对应路由
	name            string
}

func (group *routerGroup) Add(name string, handleFunc HandlerFunc) {
	group.handleMap[name] = handleFunc
}
func (group *routerGroup) Any(name string, handleFunc HandlerFunc) {
	group.handleMap[name] = handleFunc
	group.handleMapMethod[ANY] = append(group.handleMapMethod[ANY], name)
}
func (group *routerGroup) Get(name string, handleFunc HandlerFunc) {
	group.handleMap[name] = handleFunc
	group.handleMapMethod[http.MethodGet] = append(group.handleMapMethod[http.MethodGet], name)
}
func (group *routerGroup) Post(name string, handleFunc HandlerFunc) {
	group.handleMap[name] = handleFunc
	group.handleMapMethod[http.MethodPost] = append(group.handleMapMethod[http.MethodPost], name)
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	group := &routerGroup{
		handleMap:       make(map[string]HandlerFunc),
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
				//支持任意请求
				routes, ok := group.handleMapMethod[ANY]
				if ok {
					for _, routeName := range routes {
						if routeName == name {
							handlerFunc(writer, request)
							return
						}
					}
				}
				// 再遍历分组下的方法集合
				routes, ok = group.handleMapMethod[method]
				if ok {
					for _, routeName := range routes {
						if routeName == name {
							handlerFunc(writer, request)
							return
						}
					}
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
