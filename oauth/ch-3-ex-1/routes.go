package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"runtime"
)

func log(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
		fmt.Println("Handler function called - " + name)
		bytes, _ := json.MarshalIndent(&r.URL, "", "\t")
		fmt.Println(string(bytes))
		request, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(request))
		h(w, r)
		fmt.Println("-----------------------------------------------")
	}
}

func index(access_token * string, scope * string, refreshToken * string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		if len(*access_token) == 0 {
			writer.Header().Set("location", "/fetch_resource")
			writer.WriteHeader(http.StatusFound)
			return
		}

		// Execute the template with a map as context
		context := map[string]string {
			"access_token": *access_token,
			"scope": *scope,
			"refresh_token": *refreshToken,
		}

		templates := template.Must(template.ParseFiles("templates/client/index.html"))
		templates.ExecuteTemplate(writer, "index.html", context)
	}
}

func error(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/client/error.html"))
	templates.ExecuteTemplate(writer, "error.html", errorMsg)
	errorMsg = ""
}

func data(writer http.ResponseWriter, request *http.Request) {
	funcMap := template.FuncMap{"toJson": toJson}
	t := template.Must(template.New("data.html").Funcs(funcMap).ParseFiles("templates/client/data.html"))
	t.ExecuteTemplate(writer, "data.html", resource)
}

func authorize(writer http.ResponseWriter, request * http.Request) {
	state = pseudo_uuid()
	values := (request.URL).Query()
	values.Set("response_type", "code")
	values.Set("client_id", client.ClientId)
	values.Set("redirect_uri", client.RedirectURI[0])
	values.Set("state", state)
	values.Set("scope", scope)
	encodedString := values.Encode()
	redirectURI := fmt.Sprintf("%s?%s", authServer.AuthorizationEndpoint, encodedString)

	httpRedirect(writer, redirectURI)
}

func callback(writer http.ResponseWriter, request * http.Request) {

	values := (request.URL).Query()

	error := values.Get("error")
	if error != "" {
		errorMsg = error
		writer.Header().Set("location", "/error")
		writer.WriteHeader(http.StatusFound)
		return
	}

	callbackState := values.Get("state")
	if callbackState != state {
		errorMsg = "State is incorrect"
		writer.Header().Set("location", "/error")
		writer.WriteHeader(http.StatusFound)
		return
	}

	apiUrl := authServer.TokenEndpoint
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", values.Get("code"))
	data.Set("redirect_uri", client.RedirectURI[0])

	uri, _ := url.ParseRequestURI(apiUrl)
	r := createPostAndEncodeValues(uri, data)
	addBearerAndContentType(r, encodeClientCredentials())
	dumpRequestStdout(r)

	resp := dumpResponseStdout(doHttpRequest(r))
	responseBody := make([]byte, resp.ContentLength)
	_, err := resp.Body.Read(responseBody)
	fmt.Println(string(responseBody))
	accessResponse := AccessResponse{}
	err = json.Unmarshal(responseBody, &accessResponse)
	if err != nil {
		fmt.Println(err)
	}

	accessToken = accessResponse.AccessToken
	scope = accessResponse.Scope

	httpRedirect(writer, "/")
}

func tryRefreshToken(writer http.ResponseWriter, request * http.Request) {

	apiUrl := authServer.TokenEndpoint
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	uri, _ := url.ParseRequestURI(apiUrl)
	r := createPostAndEncodeValues(uri, data)
	addBearerAndContentType(r, encodeClientCredentials())
	dumpRequestStdout(r)

	resp := dumpResponseStdout(doHttpRequest(r))
	responseBody := make([]byte, resp.ContentLength)
	_, err := resp.Body.Read(responseBody)
	accessResponse := AccessResponse{}

	err = json.Unmarshal(responseBody, &accessResponse)
	if err != nil {
		fmt.Println(err)
	}

	accessToken = accessResponse.AccessToken
	scope = accessResponse.Scope
	refreshToken = accessResponse.RefreshToken

	httpRedirect(writer, "/fetch_resource")
}

func fetch_resource(writer http.ResponseWriter, request * http.Request) {

	if len(accessToken) == 0 {
		httpRedirect(writer, "authorize")
		return
	}

	r := makeAuthHeaderForProtectedResource()
	dumpRequestStdout(r)
	resp := dumpResponseStdout(doHttpRequest(r))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		displayProtectedResource(resp, writer)
		return
	} else if resp.StatusCode == 401 { // token expire or invalid
		accessToken = "" // invalidate accessToken
		if len(refreshToken) > 0 {
			tryRefreshToken(writer, request)
			return
		}
	}

	errorMsg = fmt.Sprintf("server returned response code: %d", resp.StatusCode)
	httpRedirect(writer, "/error")
}

func displayProtectedResource(resp *http.Response, writer http.ResponseWriter) {
	responseBody := make([]byte, resp.ContentLength)
	resp.Body.Read(responseBody)
	err := json.Unmarshal(responseBody, &resource)
	if err != nil {
		fmt.Println(err)
	}
	httpRedirect(writer, "/data")
}
