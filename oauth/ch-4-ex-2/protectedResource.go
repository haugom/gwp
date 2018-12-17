package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID string `json:"client_id"`
	Scope []string `json:"scope"`
}

type Words struct {
	Words []string `json:"words"`
	Timestamp string `json:"timestamp"`
}

type AppData struct {
	Words Words
	AccessToken *accessToken
}

type scope []string

type accessToken struct {
	accessToken string
	Scope       scope
}

func main() {


	appData := AppData{}
	indexHandler := http.HandlerFunc(index)
	wordsHandler := http.HandlerFunc(appData.processWordCommand)
	accessTokenHandler := accessToken{"", scope{}}
	appData.AccessToken = &accessTokenHandler

	stdChain := alice.New(myLoggingHandler, dumpRequest)
	protectedChain := alice.New(
		myLoggingHandler,
		dumpRequest,
		jsonHeaders,
		accessTokenHandler.getAccessToken,
		accessTokenHandler.validateToken,
		)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/words", protectedChain.Then(wordsHandler))

	server := http.Server{
		Addr: "127.0.0.1:9002",
		Handler:        mux,
	}
	server.ListenAndServe()
}

func jsonHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
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

func (c *accessToken) getAccessToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("authorization")
		fmt.Println(auth)
		if len(auth) > 0 && strings.Index(strings.ToLower(auth), "bearer") == 0 {
			c.accessToken = auth[len("bearer "):len(auth)]
		} else if len(r.FormValue("access_token")) > 0 {
			c.accessToken = r.FormValue("access_token")
		} else if len(r.URL.Query().Get("access_token")) > 0 {
			c.accessToken = r.URL.Query().Get("access_token")
		}
		h.ServeHTTP(w, r)
	})
}

func (c *accessToken) validateToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("/home/haugom/src/oauth-in-action-code/exercises/ch-4-ex-2/database.nosql")
		if err != nil {
			fmt.Println("Error opening JSON file:", err)
			return
		}
		defer jsonFile.Close()
		scanner := bufio.NewScanner(jsonFile)
		token := Token{}
		for scanner.Scan() {
			data := []byte(scanner.Text())
			json.Unmarshal(data, &token)
			if token.AccessToken == c.accessToken {
				break
			}
		}
		if token.AccessToken == c.accessToken {
			c.Scope = token.Scope
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

	})
}

type processWordRequest func(writer http.ResponseWriter, request *http.Request)

func (c *AppData) requireScope(h processWordRequest, scope string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isScopeRead := Contains(c.AccessToken.Scope, scope)
		if isScopeRead {
			log.Println("doing request")
			h(w, r)
		} else {
			log.Println("forbidden")
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer realm=localhost:9002, error=\"insufficient_scope\", scope=\"%s\"", scope))
			w.WriteHeader(http.StatusForbidden)
			return
		}
	})
}

func index(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/protectedResource/index.html"))
	templates.ExecuteTemplate(writer, "index.html", nil)
}

func (c *AppData) processWordCommand(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		c.requireScope(c.get, "read").ServeHTTP(writer,request)
	case "POST":
		c.requireScope(c.post, "write").ServeHTTP(writer,request)
	case "DELETE":
		c.requireScope(c.delete, "delete").ServeHTTP(writer,request)
	}
}

func (c *AppData) get(writer http.ResponseWriter, request *http.Request) {
	time.Now()
	now := time.Now()
	words := Words{Words: c.Words.Words, Timestamp: now.String()}
	output, _ := json.Marshal(&words)
	writer.Write(output)
}

func (c *AppData) post(writer http.ResponseWriter, request *http.Request) {
	word := request.FormValue("word")
	c.Words.Words = append(c.Words.Words, word)
	writer.WriteHeader(http.StatusCreated)
}

func (c *AppData) delete(writer http.ResponseWriter, request *http.Request) {
	length := len(c.Words.Words)
	if length > 0 {
		i := length-1
		c.Words.Words = append(c.Words.Words[:i], c.Words.Words[i+1:]...)
	}
	writer.WriteHeader(http.StatusNoContent)
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
