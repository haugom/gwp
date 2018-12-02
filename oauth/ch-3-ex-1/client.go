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

type AccessResponse struct {
	AccessToken 		string `json:"access_token"`
	TokenType 			string `json:"token_type"`
	Scope 				string `json:"scope"`
	RefreshToken		string `json:"refresh_token"`
}

var authServer AuthServer
var accessToken string
var scope string
var refreshToken string
var protectedResourceUrl string
var resource ProtectedResource
var client Client
var state string
var errorMsg string

func main() {

	authServer = AuthServer{
		AuthorizationEndpoint: "http://localhost:9001/authorize",
		TokenEndpoint: "http://localhost:9001/token",
	}
	accessToken = "987tghjkiu6trfghjuytrghj"
	scope = "foo bar"
	refreshToken = "j2r3oj32r23rmasd98uhjrk2o3i"
	errorMsg = ""
	protectedResourceUrl = "http://localhost:9002/resource"
	resource = ProtectedResource{}
	client = Client{
		ClientId: "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURI: []string{"http://localhost:9000/callback"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index(&accessToken, &scope, &refreshToken))
	mux.HandleFunc("/error", error)
	mux.HandleFunc("/data", data)
	mux.HandleFunc("/authorize", log(authorize))
	mux.HandleFunc("/callback", log(callback))
	mux.HandleFunc("/fetch_resource", log(fetch_resource))

	server := http.Server{
		Addr: "127.0.0.1:9000",
		Handler:        mux,
	}
	server.ListenAndServe()
}

