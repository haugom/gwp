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

var sharedAccessTokenDatabase string

type Token struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID string `json:"client_id"`
	Scope []string `json:"scope"`
	Username string `json:"user"`
}

type AppData struct {
	AccessToken *accessToken
}

type scope []string

type accessToken struct {
	accessToken string
	Scope       scope
	Username	string
}

type movies []string
type foods []string
type music []string

type Favorites struct {
	Movies movies `json:"movies"`
	Foods foods `json:"foods"`
	Music music `json:"music"`
}

type UserObject struct {
	User string `json:"user"`
    Favorites Favorites `json:"favorites"`
}

var aliceFavorites = Favorites{
	Movies:movies{"The Multidmensional Vector", "Space Fights", "Jewelry Boss"},
	Foods:foods{"bacon", "pizza", "bacon pizza"},
	Music:music{"techno", "industrial", "alternative"},
}

var bobFavorites = Favorites{
	Movies:movies{"An Unrequited Love", "Several Shades of Turquoise", "Think Of The Children"},
	Foods:foods{"bacon", "kale", "gravel"},
	Music:music{"baroque", "ukulele", "baroque ukulele"},
}

func main() {

	sharedAccessTokenDatabase = "/home/haugom/src/oauth-in-action-code/exercises/ch-4-ex-4/database.nosql"

	appData := AppData{}
	indexHandler := http.HandlerFunc(index)
	favoritesHandler := http.HandlerFunc(appData.favorites)

	accessTokenHandler := accessToken{"", scope{}, ""}
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
	mux.Handle("/favorites", protectedChain.Then(favoritesHandler))

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
		jsonFile, err := os.Open(sharedAccessTokenDatabase)
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
			c.Username = token.Username
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

func (c *AppData) favorites(writer http.ResponseWriter, request *http.Request) {
	var user UserObject
	if c.AccessToken.Username == "alice" {
		user = UserObject{"Alice", aliceFavorites}
	} else if c.AccessToken.Username == "bob" {
		user = UserObject{"Bob", bobFavorites}
	} else {
		user = UserObject{User: "Unknown"}
	}
	output, _ := json.Marshal(&user)
	writer.Write(output)
}
