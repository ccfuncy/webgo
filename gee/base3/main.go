package main

import (
	"fmt"
	"gee/gee"
	"net/http"
)

func main() {
	g := gee.New()

	g.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "URL.Path=%q\n", r.URL.Path)
	})
	g.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	g.Run(":9999")
}
