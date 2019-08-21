package main

import (
	"encoding/base64"
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
	"strings"
	"time"
)

var sharedAccessTokenDatabase string

type redirectURIS []string

type client struct {
	Name         string       `json:"name"`
	Scope        string       `json:"scope"`
	ClientID     string       `json:"client_id"`
	ClientSecret string       `json:"client_secret"`
	RedirectURIS redirectURIS `json:"redirect_uris"`
	LogoURI      string       `json:"logo_uri"`
}

type clientMap map[string]client

var allClients clientMap

type RequestMap map[string]url.Values

var requests RequestMap

type AccessCode struct {
	Code     string
	ClientID string
}

type CodeMap map[string]AccessCode

var codes CodeMap

type RefreshTokenMap map[string]RefreshTokenResponse

var refreshTokens RefreshTokenMap

type AccessTokenMap map[string]AccessTokenResponse

var accessTokens AccessTokenMap

type appData struct {
	clients *clientMap
}

type myError struct {
	error string `json:"error"`
}

func (e *myError) Error() string {
	return e.error
}

func New(text string) error {
	return &myError{text}
}

type AuthData struct {
	Client client
	ReqId  string
	Scopes []string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json"token_type"`
	Expires      int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
}

type AccessTokenResponse struct {
	AccessToken string
	ClientID    string
}

type AuthServer struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
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

	sharedAccessTokenDatabase = "/home/haugom/src/oauth-in-action-code/exercises/ch-6-ex-1/database.nosql"

	firstClient := client{
		"oauth client 1",
		"foo bar",
		"oauth-client-1",
		"oauth-client-secret-1",
		redirectURIS{"http://localhost:9000/callback"},
		"",
	}
	secondClient := client{
		"oauth client 2",
		"dex pex",
		"oauth-client-2",
		"oauth-client-secret-2",
		redirectURIS{"http://localhost:9000/callback", "http://localhost:9000/callback2"},
		"",
	}
	allClients = make(clientMap, 2)
	allClients["oauth-client-1"] = firstClient
	allClients["oauth-client-2"] = secondClient
	log.Println(allClients)

	requests = make(RequestMap, 10)
	codes = make(CodeMap, 10)
	refreshTokens = make(RefreshTokenMap, 10)
	accessTokens = make(AccessTokenMap, 10)

	appData := appData{clients: &allClients}

	log.Println(appData.clients)

	indexHandler := http.HandlerFunc(index)
	authorizeHandler := http.HandlerFunc(appData.authorize)
	approveHandler := http.HandlerFunc(appData.approve)
	tokenHandler := http.HandlerFunc(appData.token)
	tokensHandler := http.HandlerFunc(tokens)
	accessTokenHandler := http.HandlerFunc(deleteAccessToken)
	refreshTokenHandler := http.HandlerFunc(deleteRefreshToken)

	stdChain := alice.New(myLoggingHandler, dumpRequest)

	mux := http.NewServeMux()
	mux.Handle("/", stdChain.Then(indexHandler))
	mux.Handle("/authorize", stdChain.Then(authorizeHandler))
	mux.Handle("/approve", stdChain.Then(approveHandler))
	mux.Handle("/token", stdChain.Then(tokenHandler))
	mux.Handle("/tokens", stdChain.Then(tokensHandler))
	mux.Handle("/access_token", stdChain.Then(accessTokenHandler))
	mux.Handle("/refresh_token", stdChain.Then(refreshTokenHandler))

	server := http.Server{
		Addr:    "127.0.0.1:9003",
		Handler: mux,
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
	authServer := AuthServer{
		"http://127.0.0.1:9001/authorization",
		"http://127.0.0.1:9001/token",
	}
	clientsArray := make([]client, 0)
	for _, value := range allClients {
		log.Printf("Adding %s\n", value)
		clientsArray = append(clientsArray, value)
	}
	a := struct {
		AuthServer AuthServer
		Clients    []client
	}{authServer, clientsArray}

	templates := template.Must(template.ParseFiles("templates/authorizationServer/index.html"))
	e := templates.ExecuteTemplate(writer, "index.html", &a)
	if e != nil {
		log.Println(e)
	}
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
	authData := AuthData{clientId, requestId, strings.Fields(clientId.Scope)}
	context := map[string]AuthData{
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
	rscopes := make([]string, 0)

	if err := request.ParseForm(); err != nil {
		// handle error
		(&myError{error: "Unexpected error: " + err.Error()}).renderError(writer, request)
		return
	}

	for key, values := range request.PostForm {
		if strings.HasPrefix(key, "scope_") {
			left := strings.TrimLeft(key, "scope_")
			if len(values[0]) > 0 && "on" == values[0] {
				rscopes = append(rscopes, left)
			}
		}
	}

	if ok == false {
		(&myError{error: "No matching authrozation request"}).renderError(writer, request)
		return
	}
	client := c.getClient(query.Get("client_id"))
	responseType := query.Get("response_type")
	state := query.Get("state")

	if len(approved) > 0 {
		if responseType == "code" {
			code := StringWithCharset(10, charset)
			codes[code] = AccessCode{code, client.ClientID}

			writer.Header().Set("location", fmt.Sprintf("%s?code=%s&state=%s", client.RedirectURIS[0], code, state))
			writer.WriteHeader(http.StatusFound)
		} else if responseType == "token" {
			m, _ := time.ParseDuration("10s")
			now := time.Now()
			now = now.Add(m)

			accessToken := StringWithCharset(10, charset)
			accessTokens[accessToken] = AccessTokenResponse{
				accessToken,
				client.ClientID,
			}
			tokenResponse := &TokenResponse{
				accessToken,
				"Bearer",
				now.Unix() * 1000,
				"",
			}
			tokenResponseBytes, _ := json.Marshal(tokenResponse)
			outfile, _ := os.OpenFile(sharedAccessTokenDatabase, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
			defer outfile.Close()
			outfile.Write(tokenResponseBytes)
			outfile.WriteString("\n")

			writer.Header().Set("location", fmt.Sprintf("%s#access_token=%s&token_type=%s&state=%s", client.RedirectURIS[0], accessToken, "Bearer", state))
			writer.WriteHeader(http.StatusFound)
		} else {
			writer.Header().Set("location", fmt.Sprintf("%s?error=unsupported_response_type", client.RedirectURIS[0]))
			writer.WriteHeader(http.StatusFound)
		}
	} else {
		writer.Header().Set("location", fmt.Sprintf("%s?error=accedd_denied", client.RedirectURIS[0]))
		writer.WriteHeader(http.StatusFound)
	}

}

func (c *appData) token(writer http.ResponseWriter, request *http.Request) {
	auth := request.Header.Get("authorization")
	var clientId string
	var clientSecret string
	var err error
	if len(auth) > 0 && strings.Index(strings.ToLower(auth), "basic") == 0 {
		encodedCredentials := auth[len("basic "):len(auth)]
		err, clientId, clientSecret = decodeCredentials(encodedCredentials)
		if err != nil {
			output, _ := json.Marshal(&myError{error: err.Error()})
			writer.Header().Set("content-type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write(output)
			return
		}
	}

	if len(request.FormValue("client_id")) > 0 {
		if len(clientId) > 0 {
			output, _ := json.Marshal(&myError{"invalid_client"})
			writer.Header().Set("content-type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write(output)
			return
		}
		clientId = request.FormValue("client_id")
		clientSecret = request.FormValue("client_secret")
	}

	client := c.getClient(clientId)
	if len(client.ClientID) == 0 {
		output, _ := json.Marshal(&myError{"invalid_client"})
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write(output)
		return
	}

	if client.ClientSecret != clientSecret {
		output, _ := json.Marshal(&myError{"invalid_client"})
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write(output)
		return
	}

	grantType := request.FormValue("grant_type")
	isAuthorizationCode := grantType == "authorization_code"
	isRefreshToken := grantType == "refresh_token"
	if !isAuthorizationCode && !isRefreshToken {
		output, _ := json.Marshal(&myError{"unsupported_grant_type"})
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(output)
		return
	}

	if isAuthorizationCode {
		code := codes[request.FormValue("code")]
		if len(code.Code) == 0 {
			output, _ := json.Marshal(&myError{"invalid_grant"})
			writer.Header().Set("content-type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(output)
			return
		}
		delete(codes, request.FormValue("code"))

		if code.ClientID != clientId {
			output, _ := json.Marshal(&myError{"invalid_grant"})
			writer.Header().Set("content-type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(output)
			return
		}

		m, _ := time.ParseDuration("10s")
		now := time.Now()
		now = now.Add(m)

		accessToken := StringWithCharset(10, charset)
		refreshToken := StringWithCharset(10, charset)
		tokenResponse := &TokenResponse{
			accessToken,
			"Bearer",
			now.Unix() * 1000,
			refreshToken,
		}
		refreshTokenResponse := &RefreshTokenResponse{
			refreshToken,
			code.ClientID,
		}
		tokenResponseBytes, _ := json.Marshal(tokenResponse)
		refreshTokenResponseBytes, _ := json.Marshal(refreshTokenResponse)
		outfile, _ := os.OpenFile(sharedAccessTokenDatabase, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		defer outfile.Close()
		outfile.Write(tokenResponseBytes)
		outfile.WriteString("\n")
		outfile.Write(refreshTokenResponseBytes)
		outfile.WriteString("\n")

		refreshTokens[refreshToken] = *refreshTokenResponse
		accessTokens[accessToken] = AccessTokenResponse{
			accessToken,
			clientId,
		}

		writer.Header().Set("content-type", "application/json")
		writer.Write(tokenResponseBytes)
	}

	if isRefreshToken {
		token := refreshTokens[request.FormValue("refresh_token")]
		if len(token.RefreshToken) > 0 {
			if token.ClientID != clientId {
				delete(refreshTokens, request.FormValue("refresh_token"))
				output, _ := json.Marshal(&myError{"invalid_grant"})
				writer.Header().Set("content-type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write(output)
				return
			}

			m, _ := time.ParseDuration("10s")
			now := time.Now()
			now = now.Add(m)

			accessToken := StringWithCharset(10, charset)
			refreshToken := token.RefreshToken
			tokenResponse := &TokenResponse{
				accessToken,
				"Bearer",
				now.Unix() * 1000,
				refreshToken,
			}
			outfile, _ := os.OpenFile(sharedAccessTokenDatabase, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
			tokenResponseBytes, _ := json.Marshal(tokenResponse)
			defer outfile.Close()
			outfile.Write(tokenResponseBytes)
			outfile.WriteString("\n")

			writer.Header().Set("content-type", "application/json")
			writer.Write(tokenResponseBytes)

		} else {
			output, _ := json.Marshal(&myError{"invalid_grant"})
			writer.Header().Set("content-type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(output)
			return
		}
	}

}

func tokens(writer http.ResponseWriter, request *http.Request) {
	tokensArray := make([]AccessTokenResponse, 0)
	for _, value := range accessTokens {
		tokensArray = append(tokensArray, value)
	}
	refreshTokensArray := make([]RefreshTokenResponse, 0)
	for _, value := range refreshTokens {
		refreshTokensArray = append(refreshTokensArray, value)
	}
	a := struct {
		AccessTokens  []AccessTokenResponse
		RefreshTokens []RefreshTokenResponse
	}{tokensArray, refreshTokensArray}

	templates := template.Must(template.ParseFiles("templates/authorizationServer/tokens.html"))
	e := templates.ExecuteTemplate(writer, "tokens.html", &a)
	if e != nil {
		log.Println(e)
	}
}

func deleteAccessToken(writer http.ResponseWriter, request *http.Request) {
	accessToken := request.URL.Query().Get("token")
	delete(accessTokens, accessToken)
	writer.Header().Set("location", "/tokens")
	writer.WriteHeader(http.StatusFound)
}

func deleteRefreshToken(writer http.ResponseWriter, request *http.Request) {
	refreshToken := request.URL.Query().Get("token")
	delete(refreshTokens, refreshToken)
	writer.Header().Set("location", "/tokens")
	writer.WriteHeader(http.StatusFound)
}

func decodeCredentials(encodedCredentials string) (error, string, string) {
	data, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return err, "", ""
	}
	credentialsAsSlice := strings.Split(string(data), ":")
	if len(credentialsAsSlice) != 2 && len(credentialsAsSlice[0]) < 1 && len(credentialsAsSlice[1]) < 1 {
		return New("invalid authorization header"), "", ""
	}
	return nil, credentialsAsSlice[0], credentialsAsSlice[1]
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
