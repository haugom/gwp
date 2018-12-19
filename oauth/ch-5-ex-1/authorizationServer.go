
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
	"net/url"
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

type RequestMap map[string]url.Values
var requests RequestMap

type CodeMap map[string]string
var codes CodeMap

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

	firstClient := client{
		"oauth client 1",
		"oauth-client-1",
		"oauth-client-secret-1",
		redirectURIS{"http://localhost:9000/callback"},
		"",
	}
	allClients = make(clientMap, 1)
	allClients["oauth-client-1"] = firstClient

	requests = make(RequestMap, 10)
	codes = make(CodeMap, 10)

	appData := appData{clients:&allClients}

	log.Println(appData.clients)

	indexHandler := http.HandlerFunc(index)
	authorizeHandler := http.HandlerFunc(appData.authorize)
	approveHandler := http.HandlerFunc(appData.approve)

	stdChain := alice.New(myLoggingHandler, dumpRequest)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/authorize", stdChain.Then(authorizeHandler))
	mux.Handle("/approve", stdChain.Then(approveHandler))

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

func (e *myError) renderError(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/authorizationServer/error.html"))
	templates.ExecuteTemplate(writer, "error.html", e.error)
}

func (c *appData) authorize(writer http.ResponseWriter, request *http.Request) {
	clientId := c.getClient(request.URL.Query().Get("client_id"))
	if len(clientId.ClientID) == 0 {
		(&myError{error: "Unknown client"}).renderError(writer, request)
		return
	} else if Contains(clientId.RedirectURIS, request.URL.Query().Get("redirect_uri")) == false {
		output, _ := json.Marshal(&myError{error: "Invalid redirect URI"})
		writer.Write(output)
		return
	}

	requestId := StringWithCharset(10, charset)
	requests[requestId] = request.URL.Query()
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

func (c *appData) approve(writer http.ResponseWriter, request *http.Request) {
	reqId := request.FormValue("reqid")
	approved := request.FormValue("approve")
	query, ok := requests[reqId]

	if ok == false {
		(&myError{error: "No matching authrozation request"}).renderError(writer, request)
		return
	}
	clientId := c.getClient(query.Get("client_id"))
	responseType := query.Get("response_type")
	state := query.Get("state")

	if len(approved) > 0 {
		if responseType == "code" {
			code := StringWithCharset(10, charset)
			codes[code] = code

			writer.Header().Set("location", fmt.Sprintf("%s?code=%s&state=%s", clientId.RedirectURIS[0], code, state))
			writer.WriteHeader(http.StatusFound)
		} else {
			writer.Header().Set("location", fmt.Sprintf("%s?error=unsupported_responste_type", clientId.RedirectURIS[0]))
			writer.WriteHeader(http.StatusFound)
		}
	} else {
		writer.Header().Set("location", fmt.Sprintf("%s?error=accedd_denied", clientId.RedirectURIS[0]))
		writer.WriteHeader(http.StatusFound)
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
