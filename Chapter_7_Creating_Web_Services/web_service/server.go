package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path"
	"strconv"
)

const (newline = "\n")

type Post struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func main() {
	server := http.Server{
		Addr: ":8081",
	}
	http.HandleFunc("/post",  handlePostRequest)
	http.HandleFunc("/post/", handleRequest)
	server.ListenAndServe()
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlePost(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
	}
}

// main handler function
func handleRequest(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case "GET":
		err = handleGet(w, r)
	case "POST":
		err = handlePost(w, r)
	case "PUT":
		err = handlePut(w, r)
	case "DELETE":
		err = handleDelete(w, r)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Retrieve a post
// GET /post/1
func handleGet(w http.ResponseWriter, r *http.Request) (err error) {
	var output []byte
	statusCode := http.StatusOK
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		return
	}
	post, err := retrieve(id)
	if err == sql.ErrNoRows {
		statusCode = http.StatusNotFound
		err = nil
	} else if err != nil {
		return
	} else {
		output, err = json.MarshalIndent(&post, "", "\t\t")
		output = append(output, newline...)
	}
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(output)
	return
}

// Create a post
// POST /post/
func handlePost(w http.ResponseWriter, r *http.Request) (err error) {
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	var post Post
	json.Unmarshal(body, &post)
	err = post.create()
	if err != nil {
		return
	}
	w.WriteHeader(200)
	return
}

// Update a post
// PUT /post/1
func handlePut(w http.ResponseWriter, r *http.Request) (err error) {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		return
	}
	post, err := retrieve(id)
	if err != nil {
		return
	}
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	json.Unmarshal(body, &post)
	err = post.update()
	if err != nil {
		return
	}
	w.WriteHeader(200)
	return
}

// Delete a post
// DELETE /post/1
func handleDelete(w http.ResponseWriter, r *http.Request) (err error) {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		return
	}
	post, err := retrieve(id)
	if err != nil {
		return
	}
	err = post.delete()
	if err != nil {
		return
	}
	w.WriteHeader(200)
	return
}
