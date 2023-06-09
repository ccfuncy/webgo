package main

import (
	"fmt"
	"log"
	"net/http"
)

type Engine struct {
}

func (e Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/":
		fmt.Fprintf(writer, "URL.PATH=%q\n", request.URL.Path)
	case "/index":
		for k, v := range request.Header {
			fmt.Fprintf(writer, "Header[%q]=%q\n", k, v)
		}
	default:
		fmt.Fprintf(writer, "404 not found %q", request.URL)
	}
}

func main() {
	engine := Engine{}
	log.Fatal(http.ListenAndServe(":9999", engine))
}
