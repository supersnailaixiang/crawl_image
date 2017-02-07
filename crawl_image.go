package main

import (
	"fmt"
	"io"
	"net/http"

	"os"

	"golang.org/x/net/html"

	"strings"
)

func main() {
	var n int
	n++
	workList := make(chan []string)
	go func() { workList <- os.Args[1:] }()

	seen := make(map[string]bool)
	for ; n > 0; n-- {
		list := <-workList
		for _, link := range list {
			if !seen[link] {
				seen[link] = true
				n++
				go func(link string) {
					workList <- crawl(link)
				}(link)
			}
		}
	}
}

// 限制协程数量30个
var tokens = make(chan int, 30)

func crawl(url string) []string {

	tokens <- 1
	list := Extract(url)
	<-tokens
	return list
}

func Extract(url string) []string {
	resp, err := http.Get(url)
	//defer resp.Body.Close()
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil
	}
	var links []string
	doc, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil
	}
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}
				link, err := resp.Request.URL.Parse(a.Val)
				if err != nil {
					continue
				}

				links = append(links, link.String())

			}
		}

		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key != "src" {
					continue
				}
				imageUrl, err := resp.Request.URL.Parse(a.Val)
				if err != nil {
					continue
				}
				fileName := subStr(imageUrl.String())
				file, err := os.Create("./image/" + fileName)
				if err != nil {
					continue
				}
				imageResp, err := http.Get(imageUrl.String())
				if err != nil {
					continue
				}

				io.Copy(file, imageResp.Body)
				imageResp.Body.Close()
				file.Close()

			}
		}
	}
	forEachNode(doc, visitNode)
	return links
}

func forEachNode(n *html.Node, f func(n *html.Node)) {
	f(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, f)
	}
}
func subStr(str string) string {
	strAttr := strings.Split(str, "/")
	return strAttr[len(strAttr)-1]
}
func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
