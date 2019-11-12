package gosvg

import (
	"strings"

	"golang.org/x/net/html"
)

func slipTag(tag string) (namespace, local string) {
	ls := strings.Split(tag, ":")
	if len(ls) == 2 {
		return ls[0], ls[1]
	}
	return "", tag
}

// NodeChildren returns the direct children of `element`.
// Skip empty text nodes if `skipBlank` is `true`.
func nodeChildren(element *html.Node, skipBlank bool) (children []*html.Node) {
	child := element.FirstChild
	for child != nil {
		if !(skipBlank && child.Type == html.TextNode && strings.TrimSpace(child.Data) == "") {
			children = append(children, child)
		}
		child = child.NextSibling
	}
	return
}
