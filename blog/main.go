package main

import (
	"fmt"
	"gofaster"
	"net/http"
)

func main() {
	engine := gofaster.New()
	g := engine.Group("hello")
	g.Add("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})
	engine.Run()
}
