package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

type Post struct {
	XMLName xml.Name `xml:"post"`
	Id      int32   `xml:"id,attr"`
	Content string   `xml:"content"`
	Author  Author   `xml:"author"`
	Xml     string   `xml:",innerxml"`
}

type SimplePost struct {
	XMLName xml.Name `xml:"post"`
	Id      int32   `xml:"id,attr"`
	Content string   `xml:"content"`
	Author  Author   `xml:"author"`
}

type Author struct {
	Id   int32 `xml:"id,attr"`
	Name string `xml:",chardata"`
}

func main() {
	xmlFile, err := os.Open("post.xml")
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return
	}
	defer xmlFile.Close()
	xmlData, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		fmt.Println("Error reading XML data:", err)
		return
	}

	var post Post
	xml.Unmarshal(xmlData, &post)
	fmt.Println(post)

	fmt.Printf("xmlName -> %s\n", post.XMLName)
	fmt.Printf("xml -> %s\n", post.Xml)
	fmt.Printf("post id -> %d\n", post.Id)
	fmt.Printf("content: %s\n", post.Content)
	fmt.Printf("author id: %d\n", post.Author.Id)
	fmt.Printf("author name: %s\n", post.Author.Name)
	fmt.Println("---")

	simplePost := SimplePost{
		Id: post.Id,
		XMLName: post.XMLName,
		Content: post.Content,
		Author: post.Author,
	}

	output, _ := xml.MarshalIndent(simplePost, "", "  ")
	fmt.Printf("%s%s\n", xml.Header, string(output))
}
