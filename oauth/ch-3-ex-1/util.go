package main

import "encoding/json"

func toJson(resource ProtectedResource) string {
	bytes, _ := json.MarshalIndent(&resource, "", "\t")
	return string(bytes)
}