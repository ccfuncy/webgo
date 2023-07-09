package main

import (
	"fmt"
	"gofaster"
	"net/http"
)

func main() {
	engine := gofaster.New()
	g := engine.Group("user")
	//g.Any("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "any hello")
	//
	//})
	//g.Get("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "get any hello")
	//})
	g.Use(func(next gofaster.HandlerFunc) gofaster.HandlerFunc {
		return func(ctx *gofaster.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
		}
	})
	g.Get("/hello/user", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "get any hello")
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
		ctx.HTMLTemplate("login.html", "", "tpl/login.html", "tpl/header.html")
		//print(err.Error())
	})
	engine.LoadTemplate("tpl/*.html")
	g.Get("/template", func(ctx *gofaster.Context) {
		ctx.Template("index.html", "")
		//print(err.Error())
	})
	engine.Run()
}
