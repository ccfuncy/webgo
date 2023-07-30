package gofaster

import (
	"fmt"
	"gofaster/config"
	"gofaster/gateway"
	fslog "gofaster/log"
	"gofaster/render"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Engine struct {
	router
	funcMap        template.FuncMap
	HTMLRender     render.HTMLRender
	pool           sync.Pool
	Logger         *fslog.Logger
	OpenGateway    bool
	gatewayConfigs []gateway.GWConfig
}

// 实现该接口，将所有请求转交给他处理分发
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	method := request.Method
	ctx := e.pool.Get().(*Context)
	ctx.W = writer
	ctx.R = request
	//先遍历分组路由
	for _, group := range e.routerGroups {
		///再遍历分组路由下的路由集合
		routerName := SubStringLast(request.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil && len(node.childrens) == 0 {
			//支持任意请求
			handle, ok := group.handleMap[node.routerName][ANY]
			if ok {
				group.methodHandle(handle, ctx)
				//handle(c)
				return
			}

			handle, ok = group.handleMap[node.routerName][method]
			if ok {
				group.methodHandle(handle, ctx)
				return
			}
			writer.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(writer, "%s %s not allowed \n", request.RequestURI, method)
			return
		}
	}
	e.pool.Put(ctx)
	writer.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(writer, "%s %s not found \n", request.RequestURI, method)
	return
}

func New() *Engine {
	engine := &Engine{router: router{}}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}
func Default() *Engine {
	engine := New()
	engine.Logger = fslog.Default()
	path, ok := config.Conf.Log["path"]
	if ok {
		engine.Logger.SetPath(path.(string))
	}
	return engine
}

func (e *Engine) SetGatewayConfig(config []gateway.GWConfig) {
	e.gatewayConfigs = config
}

func (e *Engine) allocateContext() any {
	return &Context{
		E: e,
	}
}
func (e *Engine) Run(addr string) {
	//for _, group := range e.routerGroups {
	//	for s, handlerFunc := range group.handleMap {
	//		http.HandleFunc("/"+group.name+s, handlerFunc)
	//	}
	//}
	http.Handle("/", e)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e.Handler())
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.HTMLRender = render.HTMLRender{Template: t}
}

func (e *Engine) LoadTemplateConfig() {
	pattern, ok := config.Conf.Template["pattern"]
	if !ok {
		e.Logger.Error("config not exist template.pattern field!")
	}
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern.(string)))
	e.HTMLRender = render.HTMLRender{Template: t}
}

func (e *Engine) Handler() http.Handler {
	return e
}
