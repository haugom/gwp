package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
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

func main() {
	protectedResource := ProtectedResource{
		Name: "Protected Resource",
		Description: "This data has been protected by OAuth 2.0",
	}


	inToken := ""

	mux := http.NewServeMux()
	mux.HandleFunc("/", log(index))
	mux.HandleFunc("/resource", log(getAccessToken(&inToken, resource(protectedResource))))

	server := http.Server{
		Addr: "127.0.0.1:9002",
		Handler:        mux,
	}
	server.ListenAndServe()
}

func log(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(request))
		h(w, r)
	}
}

func getAccessToken(inToken *string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenValid := false
		auth := r.Header.Get("authorization")
		fmt.Println(auth)
		if len(auth) > 0 && strings.Index(strings.ToLower(auth), "bearer") == 0 {
			*inToken = auth[len("bearer "):len(auth)]
		} else if len(r.FormValue("access_token")) > 0 {
			*inToken = r.FormValue("access_token")
		} else if len(r.URL.Query().Get("access_token")) > 0 {
			*inToken = r.URL.Query().Get("access_token")
		}
		validateToken(&tokenValid, inToken)
		if tokenValid == true {
			h(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func validateToken(valid *bool, inToken *string) {
	jsonFile, err := os.Open("/home/haugom/src/oauth-in-action-code/exercises/ch-4-ex-1/database.nosql")
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer jsonFile.Close()
	scanner := bufio.NewScanner(jsonFile)
	for scanner.Scan() {
		token := Token{}
		data := []byte(scanner.Text())
		json.Unmarshal(data, &token)
		if token.AccessToken == *inToken {
			*valid = true
			return
		}
	}
}

func index(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/protectedResource/index.html"))
	templates.ExecuteTemplate(writer, "index.html", nil)
}

func resource(protectedResource ProtectedResource) http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {
		output, err := json.MarshalIndent(&protectedResource, "", "\t")
		output = append(output, newline...)
		if err != nil {
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(output)
	}
}
