package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func makeAuthHeaderForProtectedResource() *http.Request {
	apiUrl := protectedResourceUrl
	data := url.Values{}
	uri, _ := url.ParseRequestURI(apiUrl)
	r, _ := http.NewRequest("POST", uri.String(), strings.NewReader(data.Encode()))
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return r
}

func doHttpRequest(r *http.Request) *http.Response {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func httpRedirect(writer http.ResponseWriter, location string) {
	writer.Header().Set("location", location)
	writer.WriteHeader(http.StatusFound)
}

func encodeClientCredentials() string {
	clientString := fmt.Sprintf("%s:%s", client.ClientId, client.ClientSecret)
	fmt.Println(clientString)
	bearer := []byte(clientString)
	encodedBearer := base64.StdEncoding.EncodeToString(bearer)
	return encodedBearer
}

func addBearerAndContentType(request *http.Request, encodedBearer string) {
	request.Header.Add("Authorization", fmt.Sprintf("Basic: %s", encodedBearer))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
}

func createPostAndEncodeValues(uri *url.URL, values url.Values) *http.Request {
	r, _ := http.NewRequest("POST", uri.String(), strings.NewReader(values.Encode()))
	r.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))
	return r
}
