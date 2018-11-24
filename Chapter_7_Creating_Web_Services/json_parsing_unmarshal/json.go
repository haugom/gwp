package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Post struct {
	Id       int       `json:"id"`
	Content  string    `json:"content"`
	NotThere string    `json:"nonexistant"`
	Author   Author    `json:"author"`
	Comments []Comment `json:"comments"`
}

type Author struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Comment struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func jsonSlurp(fileName string) (jsonData []byte, err error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return nil, err
	}
	defer jsonFile.Close()
	jsonData, err = ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading JSON data:", err)
		return nil, err
	}
	return jsonData, nil
}

func main() {
	jsonData, err := jsonSlurp("post.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonData))
	var post Post
	err = json.Unmarshal(jsonData, &post)
	if err != nil {
		fmt.Println("Error unmarshal json data", err)
		return
	}
	fmt.Println(post.Id)
	fmt.Println(post.Content)
	fmt.Println(post.NotThere)
	fmt.Println(post.Author.Id)
	fmt.Println(post.Author.Name)
	fmt.Println(post.Comments[0].Id)
	fmt.Println(post.Comments[0].Content)
	fmt.Println(post.Comments[0].Author)

}
