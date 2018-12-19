
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)
type redirectURIS []string

type client struct {
	Name string `json:"name"`
	ClientID string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURIS redirectURIS `json:"redirect_uris"`
	LogoURI string `json:"logo_uri"`
}

type clientMap map[string]client

var allClients clientMap

type RequestMap map[string]string

var requests RequestMap

type appData struct {
	clients *clientMap
}

type myError struct {
	error string `json:"error"`
}

type AuthData struct {
	Client client
	ReqId string
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func main() {

	allClients = clientMap{
		"oauth-client-1": {
			"oauth client 1",
			"oauth-client-1",
			"oauth-client-secret-1",
			redirectURIS{"http://localhost:9000/callback"},
			"",
		},
	}

	requests = make(RequestMap, 10)

	appData := appData{clients:&allClients}

	log.Println(appData.clients)

	indexHandler := http.HandlerFunc(index)
	errorHandler := http.HandlerFunc(error)
	authorizeHandler := http.HandlerFunc(appData.authorize)

	stdChain := alice.New(myLoggingHandler, dumpRequest)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/error", stdChain.Then(errorHandler))
	mux.Handle("/authorize", stdChain.Then(authorizeHandler))

	server := http.Server{
		Addr: "127.0.0.1:9001",
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

func (c *appData) authorize(writer http.ResponseWriter, request *http.Request) {
	clientId := c.getClient(request.URL.Query().Get("client_id"))
	if len(clientId.ClientID) == 0 {
		output, _ := json.Marshal(&myError{error: "Unknown client"})
		writer.Write(output)
		return
	} else if Contains(clientId.RedirectURIS, request.URL.Query().Get("redirect_uri")) == false {
		output, _ := json.Marshal(&myError{error: "Invalid redirect URI"})
		writer.Write(output)
		return
	}

	requestId := StringWithCharset(10, charset)
	requests[requestId] = requestId
	authData := AuthData{clientId, requestId}
	context := map[string]AuthData {
		"auth": authData,
	}

	templates := template.Must(template.ParseFiles("templates/authorizationServer/approve.html"))
	error := templates.ExecuteTemplate(writer, "approve.html", context)
	if error != nil {
		log.Println(error)
	}
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
