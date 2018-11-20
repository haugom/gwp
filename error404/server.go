package main

import (
	"html/template"
	"net/http"
	"time"
)

func error(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	t, _ := template.ParseFiles("templates/404.html")
	t.Execute(writer, "")
}

func main() {
	p("Custom error handler4fun", version(), "started at", config.Address)

	// handle static assets
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(config.Static))
	mux.Handle("/images/", http.StripPrefix("/images/", files))
	mux.HandleFunc("/", error)

	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
}