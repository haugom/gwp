package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

const (newline = "\n")

type ProtectedResource struct {
	Name string `json:"name"`
	Description string `json:"description"`
}

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

type accessToken struct {
	token string
}

func main() {

	words := Words{}

	indexHandler := http.HandlerFunc(index)
	wordsHandler := http.HandlerFunc(words.processWordCommand)
	accessTokenHandler := accessToken{""}

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
			c.token = auth[len("bearer "):len(auth)]
		} else if len(r.FormValue("access_token")) > 0 {
			c.token = r.FormValue("access_token")
		} else if len(r.URL.Query().Get("access_token")) > 0 {
			c.token = r.URL.Query().Get("access_token")
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
			if token.AccessToken == c.token {
				break
			}
		}
		if token.AccessToken == c.token {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

	})
}

func index(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/protectedResource/index.html"))
	templates.ExecuteTemplate(writer, "index.html", nil)
}

func (c *ProtectedResource) resource(writer http.ResponseWriter, request *http.Request) {
	output, err := json.MarshalIndent(&c, "", "\t")
	output = append(output, newline...)
	if err != nil {
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(output)
}

func (c *Words) processWordCommand(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		c.get(writer, request)
	case "POST":
		c.post(writer, request)
	case "DELETE":
		c.delete(writer, request)
	}
}

func (c *Words) get(writer http.ResponseWriter, request *http.Request) {
	time.Now()
	now := time.Now()
	words := Words{Words: c.Words, Timestamp: now.String()}
	output, _ := json.Marshal(&words)
	writer.Write(output)
}

func (c *Words) post(writer http.ResponseWriter, request *http.Request) {
	word := request.FormValue("word")
	c.Words = append(c.Words, word)
	writer.WriteHeader(http.StatusCreated)
}

func (c *Words) delete(writer http.ResponseWriter, request *http.Request) {
	length := len(c.Words)
	if length > 0 {
		i := length-1
		c.Words = append(c.Words[:i], c.Words[i+1:]...)
	}
	writer.WriteHeader(http.StatusNoContent)
}

