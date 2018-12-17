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
)

type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID string `json:"client_id"`
	Scope []string `json:"scope"`
}

type AppData struct {
	AccessToken *accessToken
}

type scope []string

type accessToken struct {
	accessToken string
	Scope       scope
}

type fruit []string
type veggies []string
type meats []string

type Catalog struct {
	Fruit fruit `json:"fruit"`
	Veggies veggies `json:"veggies"`
	Meats meats `json:"meats"`
}

func main() {

	appData := AppData{}
	indexHandler := http.HandlerFunc(index)
	produceHandler := http.HandlerFunc(appData.produce)

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
	mux.Handle("/produce", protectedChain.Then(produceHandler))

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
		jsonFile, err := os.Open("/home/haugom/src/oauth-in-action-code/exercises/ch-4-ex-3/database.nosql")
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

func index(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/protectedResource/index.html"))
	templates.ExecuteTemplate(writer, "index.html", nil)
}

func (c *AppData) produce(writer http.ResponseWriter, request *http.Request) {
	fruits := fruit{"apple", "banana", "kiwi"}
	veggies := veggies{"lettuce", "onion", "potato"}
	meats := meats{"bacon", "steak", "chicken breast"}
	catalog := Catalog{}

	if Contains(c.AccessToken.Scope, "fruit") {
		catalog.Fruit = fruits
	}
	if Contains(c.AccessToken.Scope, "veggies") {
		catalog.Veggies = veggies
	}
	if Contains(c.AccessToken.Scope, "meats") {
		catalog.Meats = meats
	}
	output, _ := json.Marshal(&catalog)
	writer.Write(output)
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
