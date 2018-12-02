package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

func toJson(resource ProtectedResource) string {
	bytes, _ := json.MarshalIndent(&resource, "", "\t")
	return string(bytes)
}

func pseudo_uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}

func dumpResponseStdout(resp * http.Response) * http.Response {
	bytes, _ := httputil.DumpResponse(resp, true)
	fmt.Println("----response--------------------")
	fmt.Println(string(bytes))
	fmt.Println("--------------------------------")
	return resp
}

func dumpRequestStdout(r *http.Request) *http.Request {
	requestBytes, _ := httputil.DumpRequestOut(r, true)
	fmt.Println(string(requestBytes))
	return r
}
