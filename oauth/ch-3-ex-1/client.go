package main

import (
	"net/http"
)

type AuthServer struct {
	AuthorizationEndpoint 	string		`json: "authorizationEndpoint"`
	TokenEndpoint			string		`json: "tokenEndpoint"`
}

type ProtectedResource struct {
	Name 			string 	`json: "name"`
	Description 	string 	`json: "description""`
}

type Client struct {
	ClientId string			`json: "client_id"`
	ClientSecret string		`json: "client_secret"`
	RedirectURI []string	`json: "redirect_uris"`
}

var authServer AuthServer
var access_token string
var scope string
var resource ProtectedResource
var client Client

func main() {

	authServer = AuthServer{
		AuthorizationEndpoint: "http://localhost:9001/authorize",
		TokenEndpoint: "http://localhost:9001/token",
	}
	access_token = ""
	scope = ""
	resource = ProtectedResource{}
	client = Client{
		ClientId: "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURI: []string{"http://localhost:9000/callback"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index(&access_token, &scope))
	mux.HandleFunc("/error", error)
	mux.HandleFunc("/data", data)
	mux.HandleFunc("/authorize", log(authorize))
	mux.HandleFunc("/callback", log(callback))

	server := http.Server{
		Addr: "127.0.0.1:9000",
		Handler:        mux,
	}
	server.ListenAndServe()
}

