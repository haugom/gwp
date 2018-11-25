package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"
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

func index(access_token * string, scope * string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Execute the template with a map as context
		context := map[string]string {
			"access_token": *access_token,
			"scope": *scope,
		}

		templates := template.Must(template.ParseFiles("templates/client/index.html"))
		templates.ExecuteTemplate(writer, "index.html", context)
	}
}

func error(writer http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("templates/client/error.html"))
	templates.ExecuteTemplate(writer, "error.html", "")
}

func data(writer http.ResponseWriter, request *http.Request) {
	funcMap := template.FuncMap{"toJson": toJson}
	t := template.Must(template.New("data.html").Funcs(funcMap).ParseFiles("templates/client/data.html"))
	t.ExecuteTemplate(writer, "data.html", resource)
}

func authorize(writer http.ResponseWriter, request * http.Request) {
	values := (request.URL).Query()
	values.Set("response_type", "code")
	values.Set("client_id", client.ClientId)
	values.Set("redirect_uri", client.RedirectURI[0])
	encodedString := values.Encode()
	redirectURI := fmt.Sprintf("%s?%s", authServer.AuthorizationEndpoint, encodedString)

	writer.Header().Set("location", redirectURI)
	writer.WriteHeader(http.StatusFound)
}

func callback(writer http.ResponseWriter, request * http.Request) {

	values := (request.URL).Query()

	error := values.Get("error")
	if error != "" {
		writer.Header().Set("location", fmt.Sprintf("/error?error=%s", error))
		writer.WriteHeader(http.StatusFound)
		return
	}

	apiUrl := authServer.TokenEndpoint
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", values.Get("code"))
	data.Set("redirect_uri", client.RedirectURI[0])

	clientString := fmt.Sprintf("%s:%s", client.ClientId, client.ClientSecret)
	fmt.Println(clientString)
	bearer := []byte(clientString)
	encodedBearer := base64.StdEncoding.EncodeToString(bearer)
	fmt.Println(encodedBearer)

	httpClient := &http.Client{}
	uri, _ := url.ParseRequestURI(apiUrl)
	r, _ := http.NewRequest("POST", uri.String(), strings.NewReader(data.Encode()))
	r.Header.Add("Authorization", fmt.Sprintf("Basic: %s", encodedBearer))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	requestBytes, _ := httputil.DumpRequestOut(r, true)
	fmt.Println(string(requestBytes))

	resp, _ := httpClient.Do(r)
	responseBody := make([]byte, resp.ContentLength)
	numberOfBytes, err := resp.Body.Read(responseBody)

	response, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("Bytes read: %d, err: %v, data read: [%s]\n", numberOfBytes, err, string(responseBody))
	fmt.Println(string(response))
	writer.WriteHeader(http.StatusOK)
}