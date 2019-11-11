package test

import (
	"fmt"
	"strings"
	"testing"

	go_locale "github.com/jmshal/go-locale"

	"golang.org/x/net/html"
)

func printTree(n *html.Node) {
	fmt.Printf("%p : %v %s \n", n, n.Type, n.Data)
	fmt.Println(n.Attr)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
			continue
		}
		printTree(c)
	}
}



func TestXML(t *testing.T) {
	s := `<svg width="12cm" height="5.25cm" viewBox="0 0 1200 400"
		 xmlns="http://www.w3.org/2000/svg" requiredAttributes xmlns:fictional="http://characters.example.com" version="1.1">
	  <fictionnal:title>Example arcs01 - arc commands in path data</fictionnal:title>
	  <desc>Picture of a pie chart with two pie wedges and
			a picture of a line with arc blips</desc>
			<path d="M300,200 h-150 a150,150 0 1,0 150,-150 z"
				  fill="red" stroke="blue" stroke-width="5"></path>
	  <fictionnal:rect x="1" y="1" width="1198" height="398"
			fill="none" stroke="blue" stroke-width="1" />
	
	  <path d="M275,175 v-150 a150,150 0 0,0 -150,150 z"
			fill="yellow" stroke="blue" stroke-width="5" />
	
	  <path d="M600,350 l 50,-25 
			   a25,25 -30 0,1 50,-25 l 50,-25 
			   a25,50 -30 0,1 50,-25 l 50,-25 
			   a25,75 -30 0,1 50,-25 l 50,-25 
			   a25,100 -30 0,1 50,-25 l 50,-25"
			fill="none" stroke="red" stroke-width="5"  />
	</svg>`
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	printTree(extractSVG(root))

	// sv, err := svgFromXML(strings.NewReader(s))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(sv.attributes, sv.Paths[0].attributes)
}

func TestLocale(t *testing.T) {
	fmt.Println(go_locale.DetectLocale())
}
