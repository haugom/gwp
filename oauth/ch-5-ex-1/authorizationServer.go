
package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
)
type redirectURIS []string

type client struct {
	ClientID string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURIS redirectURIS `json:"redirect_uris"`
}

var Clients []client

func main() {

	Clients = append(Clients,
		client{
			"oauth-client-1",
			"oauth-client-secret-1",
			redirectURIS{"http://localhost:9000/callback"},
		})

	indexHandler := http.HandlerFunc(index)
	errorHandler := http.HandlerFunc(error)
	approveHandler := http.HandlerFunc(approve)

	stdChain := alice.New(myLoggingHandler, dumpRequest)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/error", stdChain.Then(errorHandler))
	mux.Handle("/approve", stdChain.Then(approveHandler))

	server := http.Server{
		Addr: "127.0.0.1:9002",
		Handler:        mux,
	}
	server.ListenAndServe()

}

func dumpRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(request))
		h.ServeHTTP(w, r)
	})
}

func myLoggingHandler(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}

func index(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/authorizationServer/index.html"))
	templates.ExecuteTemplate(writer, "index.html", nil)
}

func error(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/authorizationServer/error.html"))
	templates.ExecuteTemplate(writer, "error.html", nil)
}

func approve(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/authorizationServer/approve.html"))
	templates.ExecuteTemplate(writer, "approve.html", nil)
}
