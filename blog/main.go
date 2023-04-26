package main

import (
	"fmt"
	"gofaster"
)

func main() {
	engine := gofaster.New()
	g := engine.Group("hello")
	g.Any("/hello", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "any hello")

	})
	g.Get("/hello", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "get any hello")

	})
	g.Get("/get", func(ctx *gofaster.Context) {
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
	engine.Run()
}
