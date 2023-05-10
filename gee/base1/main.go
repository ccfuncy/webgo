package main

import (
	"fmt"
	"log"
	"net/http"
)

type Test struct {
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/index", indexHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}
func helloHandler(writer http.ResponseWriter, request *http.Request) {
	for k, v := range request.Header {
		fmt.Fprintf(writer, "Header[%q]=%q\n", k, v)
	}
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Url.Path = %q\n", request.URL.Path)
}
