
package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"html/template"
	"log"
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

type clients map[string]client

type appData struct {
	clients * clients
}

var myClients clients

type myError struct {
	error string `json:"error"`
}

var currentError myError

func main() {

	currentError = myError{}
	myClients = clients{
		"oauth-client-1": {
			"oauth-client-1",
			"oauth-client-secret-1",
			redirectURIS{"http://localhost:9000/callback"},
		},
	}

	appData := appData{clients:&myClients}

	log.Println(appData.clients)

	indexHandler := http.HandlerFunc(index)
	errorHandler := http.HandlerFunc(error)
	approveHandler := http.HandlerFunc(approve)
	authorizeHandler := http.HandlerFunc(appData.authorize)

	stdChain := alice.New(myLoggingHandler, dumpRequest)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/error", stdChain.Then(errorHandler))
	mux.Handle("/approve", stdChain.Then(approveHandler))
	mux.Handle("/authorize", stdChain.Then(authorizeHandler))

	server := http.Server{
		Addr: "127.0.0.1:9002",
		Handler:        mux,
	}
	server.ListenAndServe()

}

func (c *appData) getClient(ClientId string) client {
	theClient, ok := (*c.clients)[ClientId]
	if ok {
		return theClient
	} else {
		return client{}
	}
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

func (c *appData) authorize(writer http.ResponseWriter, request *http.Request) {
	clientId := c.getClient(request.URL.Query().Get("client_id"))
	if len(clientId.ClientID) == 0 {
		request.Context()
	}

}
