package main

import (
	"fmt"
	"net/http"
)

func process(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	//fmt.Fprintln(w, r.Form)
	//fmt.Fprintln(w, r.PostForm)
	//r.ParseMultipartForm(1024)
	//fmt.Fprintln(w, r.MultipartForm)

	fmt.Fprintln(w, "(1)", r.FormValue("hello"))
	fmt.Fprintln(w, "(2)", r.PostFormValue(("hello")))
	fmt.Fprintln(w, "(3)", r.PostForm)
	fmt.Fprintln(w, "(4)", r.MultipartForm)
}

func main() {
	server := http.Server{
		Addr: "127.0.0.1:8081",
	}
	http.HandleFunc("/process", process)
	server.ListenAndServe()
}
