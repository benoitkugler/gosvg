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

func extractText(n *html.Node) string {
	if c := n.FirstChild; c != nil && c.Type == html.TextNode {
		return c.Data
	}
	return ""
}

func extractTail(n *html.Node) string {
	if c := n.LastChild; c != nil && c.Type == html.TextNode {
		return c.Data
	}
	return ""
}

func getAttr(h *html.Node, name string) string {
	for _, attr := range h.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
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
