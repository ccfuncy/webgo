package main

import (
	"fmt"
	"gofaster"
	"gofaster/log"
	"net/http"
)

func main() {
	engine := gofaster.New()
	g := engine.Group("user")
	g.Use(gofaster.Logging)
	//g.Any("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "any hello")
	//
	//})
	//g.Get("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "get any hello")
	//})
	g.Use(func(next gofaster.HandlerFunc) gofaster.HandlerFunc {
		return func(ctx *gofaster.Context) {
			//fmt.Println("pre handler")
			next(ctx)
			//fmt.Println("post handler")
		}
	})
	g.Get("/hello/user", func(ctx *gofaster.Context) {
		logger := log.Default()
		logger.SetPath("./log")
		logger.LogFileSize = 1 << 10 //1k
		logger.Formatter = log.JsonFormatter{}
		logger.Info("12")
		logger.Error("123")
		logger.WithFields(log.Fields{"name": "ccfuncy", "id": 2}).Debug("321")
		ctx.Template("login.html", 12)
	})
	g.Get("/get", func(ctx *gofaster.Context) {
		fmt.Println("run")
		fmt.Fprint(ctx.W, "get hello")
	})
	g.Post("/post", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "post hello")
	})
	//g.Get("/use/:id", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "/use/:id get hello")
	//})
	g.Get("/use/*/get:id", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "/use/*/get:id get hello")
	})

	g.Get("/html", func(ctx *gofaster.Context) {
		ctx.HTML(http.StatusOK, "<h1>Hello</h1>")
	})
	g.Get("/htmlTemplate", func(ctx *gofaster.Context) {
		ctx.Template("login.html", "")
		//print(err.Error())
	})
	engine.LoadTemplate("tpl/*.html")
	g.Get("/template", func(ctx *gofaster.Context) {
		ctx.Template("index.html", "")
		//print(err.Error())
	})
	engine.Run()
}
