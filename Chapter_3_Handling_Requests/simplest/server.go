package main

import (
	"fmt"
	"net/http"
)

type HelloHandler struct{}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

type WorldHandler struct{}

func (h *WorldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "World!")
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Go!")
}

func programming(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "programming")
}

func main() {
	hello := HelloHandler{}
	world := WorldHandler{}
	programming := http.HandlerFunc(programming)
	server := http.Server{
		Addr: "127.0.0.1:8081",
	}
	http.Handle("/hello", &hello)
	http.Handle("/world", &world)
	http.Handle("/programming", programming)
	http.HandleFunc("/go", goHandler)
	server.ListenAndServe()
}
