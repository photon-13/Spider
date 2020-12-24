package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

//Node item
type Node struct {
	id      int
	link    string
	depth   int
	parent  *Node
	visited bool
	host    string
	scheme  string
	path    string
}

//NewNode node
func NewNode(link string, parent *Node) *Node {
	node := &Node{}
	if !strings.HasPrefix(link, "http") {
		if strings.HasPrefix(link, "/") {
			node.link = parent.scheme + "://" + parent.host + link
		} else {
			node.link = parent.scheme + "://" + parent.host + "/" + link
		}
		node.scheme = parent.scheme
		node.host = parent.host
	} else {
		node.link = link
		u, err := url.Parse(link)
		if err != nil {
			log.Println(err)
		}
		node.scheme = u.Scheme
		node.host = u.Host
	}
	if parent != nil {
		node.depth = parent.depth + 1
	}

	return node
}

//Graph queue
type Graph struct {
	queue []*Node
	graph map[string]*Node
}

//NewGraph graph
func NewGraph() *Graph {
	g := &Graph{}
	g.graph = make(map[string]*Node)
	return g
}

func (g *Graph) hasNext() bool {
	i := len(g.queue)
	if i == 0 {
		return false
	}
	return true
}

func (g *Graph) add(node *Node) {
	if _, ok := g.graph[node.link]; ok {
		return
	}
	g.queue = append(g.queue, node)
	g.graph[node.link] = node
}

func (g *Graph) get() *Node {
	node := g.queue[0]
	node.visited = true
	g.queue = g.queue[1:]
	return node
}

func getLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						links = append(links, link)
					}
				}
			}
		}
	}
}

func crawl(g *Graph) {
	for g.hasNext() {
		pnode := g.get()
		fmt.Println("Starting for link", pnode.link)
		resp, err := http.Get(pnode.link)
		if err != nil {
			log.Println(err)
		}
		for _, child := range getLinks(resp.Body) {
			cnode := NewNode(child, pnode)
			g.add(cnode)
		}
	}
}

func main() {

	g := NewGraph()
	link := "http://golang.org"
	node := NewNode(link, nil)

	g.add(node)
	crawl(g)
}
