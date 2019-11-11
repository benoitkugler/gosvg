package test

import (
	"strings"

	"golang.org/x/net/html"
)

type SVGNode struct {
	*html.Node

	namespaces map[string]string
}

func parseNamespaces(attrs []html.Attribute) map[string]string {
	out := map[string]string{}
	for _, attr := range attrs {
		if strings.Contains(attr.Key, "xmlns:") {
			out[strings.TrimPrefix(attr.Key, "xmlns:")] = attr.Val
		}
	}
	return out
}

func NewSVGNode(node *html.Node) SVGNode {
	s := SVGNode{Node: node}
	s.namespaces = parseNamespaces(node.Attr)
	return s
}
