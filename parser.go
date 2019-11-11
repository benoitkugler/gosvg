package gosvg

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"

	"golang.org/x/net/html"
)

var (
	// "display" is actually inherited but handled differently because some markers
	// are part of a none-displaying group (see test painting-marker-07-f.svg)
	NOTINHERITEDATTRIBUTES = utils.NewSet(
		"clip",
		"clip-path",
		"display",
		"filter",
		"height",
		"id",
		"mask",
		"opacity",
		"overflow",
		"rotate",
		"stop-color",
		"stop-opacity",
		"style",
		"transform",
		"viewBox",
		"width",
		"x",
		"y",
		"dx",
		"dy",
		"{http://www.w3.org/1999/xlink}href",
		"href",
	)

	COLORATTRIBUTES = utils.NewSet(
		"fill",
		"flood-color",
		"lighting-color",
		"stop-color",
		"stroke",
	)

	caseSensitiveStyleMethods = map[string]func(string) string{
		"id":            normalizeNoopStyleDeclaration,
		"class":         normalizeNoopStyleDeclaration,
		"font-family":   normalizeNoopStyleDeclaration,
		"font":          normalizeFontStyleDeclaration,
		"clip-path":     normalizeUrlStyleDeclaration,
		"color-profile": normalizeUrlStyleDeclaration,
		"cursor":        normalizeUrlStyleDeclaration,
		"fill":          normalizeUrlStyleDeclaration,
		"filter":        normalizeUrlStyleDeclaration,
		"marker-start":  normalizeUrlStyleDeclaration,
		"marker-mid":    normalizeUrlStyleDeclaration,
		"marker-end":    normalizeUrlStyleDeclaration,
		"mask":          normalizeUrlStyleDeclaration,
		"stroke":        normalizeUrlStyleDeclaration,
	}

	replacerPresere     = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	replacerNotPreserve = strings.NewReplacer("\n", "", "\r", "", "\t", " ", " +", " ")

	regexStyle = regexp.MustCompile("(?i)(.*?)" + // non-URL part (will be normalized)
		"(?:" +
		`url\(\s*` + // url(<whitespace>
		"(?:" +
		`"(?:\\.|[^"])*"` + // "<url>"
		`| \'(?:\\.|[^\'])*\'` + // '<url>'
		`| (?:\\.|[^\)])*` + // <url>
		")" +
		`\s*\)` + // <whitespace>)
		"|$" +
		")")

	regexFont = regexp.MustCompile(`^(` +
		`(\d[^\s,]*|\w[^\s,]*)` + // <size>, <length> or <identifier>
		`(\s+|\s*,\s*)` + // <whitespace> and/or comma
		`)*` + // Repeat until last
		`\d[^\s,]*`) // <size> or <line-height>
)

// Handle white spaces in text nodes.
// See http://www.w3.org/TR/SVG/text.html#WhiteSpace
func handleWhiteSpaces(s string, preserve bool) string {
	if s == "" {
		return ""
	}
	if preserve {
		return replacerPresere.Replace(s)
	} else {
		return replacerNotPreserve.Replace(s)
	}
}

// Normalize style declaration consisting of name/value pair.
// Names are always case insensitive, make all lowercase.
// Values are case insensitive in most cases. Adapt for "specials":
//     id - case sensitive identifier
//     class - case sensitive identifier(s)
//     font-family - case sensitive name(s)
//     font - shorthand in which font-family is case sensitive
//     any declaration with url in value - url is case sensitive
func normalizeStyleDeclaration(name, value string) (string, string) {
	name = strings.ToLower(strings.TrimSpace(name))
	value = strings.TrimSpace(value)
	if f, in := caseSensitiveStyleMethods[name]; in {
		value = f(value)
	} else {
		value = strings.ToLower(value)
	}
	return name, value
}

// No-operation for normalization where value is case sensitive.
// This is actually the exception to the rule. Normally value will be made
// lowercase (see normalizeStyleDeclaration above).
func normalizeNoopStyleDeclaration(value string) string {
	return value
}

// Normalize style declaration, but keep URL's as-is.
// Lowercase everything except for the URL.
func normalizeUrlStyleDeclaration(value string) string {
	for _, match := range regexStyle.FindAllStringSubmatchIndex(value, -1) {
		start, _ := match[0], match[1]
		groupStart, groupEnd := match[2], match[3]
		valueStart := ""
		if start > 0 {
			valueStart = value[:start]
		}
		normalizedValue := strings.ToLower(value[groupStart:groupEnd])
		valueEnd := value[start+len(normalizedValue):]
		value = valueStart + normalizedValue + valueEnd
	}
	return value
}

// Make first part of font style declaration lowercase (case insensitive).
//  Lowercase first part of declaration. Only the font name is case sensitive.
//  The font name is at the end of the declaration && can be "recognized"
//  by being preceded by a size or line height. There can actually be multiple
//  names. So the first part is "calculated" by selecting everything up to and
//  including the last valid token followed by a size or line height (both
//  starting with a number). A valid token is either a size/length or an
//  identifier.
//  See http://www.w3.org/TR/css-fonts-3/#font-prop
func normalizeFontStyleDeclaration(value string) string {
	return regexFont.ReplaceAllStringFunc(value, strings.ToLower)
}

type UrlFectcher func(url, mimeType string) ([]byte, error)

// Node is a SVG node.
type Node struct {
	element  *html.Node
	parent   *Node
	children []*Node
	root     bool

	namespaces map[string]string
	style      styleMatchers
	tag        string
	text       string
	urlFetcher UrlFectcher
	url        *url.URL

	attributes map[string]string
}

type styleMatchers struct {
	normal, important matcher
}

// Create the Node from ElementTree ``node``, with ``parent`` Node.
// parent=None, parentChildren=false, url=None
func NewNode(element *html.Node, style styleMatchers, urlFetcher UrlFectcher,
	parent *Node, parentChildren bool, url *url.URL) *Node {
	self := new(Node)

	self.element = element
	self.style = style

	namespace, tag := slipTag(element.Data)
	namespaceUrl := parent.namespaces[namespace]
	if namespace == "" || namespaceUrl == "" {
		self.tag = tag
	} else {
		self.tag = fmt.Sprintf("{%s}%s", namespaceUrl, tag)
	}

	self.text = extractText(element)
	self.urlFetcher = urlFetcher

	self.attributes = map[string]string{}
	self.parent = parent
	// Inherits from parent properties
	if parent != nil {
		for attribute, parentValue := range parent.attributes {
			if !NOTINHERITEDATTRIBUTES.Has(attribute) {
				self.attributes[attribute] = parentValue
			}
		}
		if url == nil {
			url = parent.url
		}
	}
	self.url = url

	// Apply CSS rules
	styleAttr := getAttr(element, "style")
	var normalAttr, importantAttr [][2]string
	if styleAttr != "" {
		normalAttr, importantAttr = parseDeclarations(styleAttr)
	}
	normalMatcher, importantMatcher := style.normal, style.important
	var allDeclarationsLists [][][2]string
	for _, rule := range normalMatcher.Match(element) {
		allDeclarationsLists = append(allDeclarationsLists, rule.payload)
	}
	allDeclarationsLists = append(allDeclarationsLists, normalAttr)
	for _, rule := range importantMatcher.Match(element) {
		allDeclarationsLists = append(allDeclarationsLists, rule.payload)
	}
	allDeclarationsLists = append(allDeclarationsLists, importantAttr)
	for _, declarations := range allDeclarationsLists {
		for _, tmp := range declarations {
			name, value := tmp[0], tmp[1]
			self.attributes[name] = strings.TrimSpace(value)
		}
	}

	// Replace currentColor by a real color value
	for attribute := range COLORATTRIBUTES {
		if self.attributes[attribute] == "currentColor" {
			c := self.attributes["color"]
			if c == "" {
				c = "black"
			}
			self.attributes[attribute] = c
		}
	}

	// Replace inherit by the parent value
	var parentAttrs map[string]string
	if parent != nil {
		parentAttrs = parent.attributes
	}
	for attribute, value := range self.attributes {
		if value == "inherit" {
			if parentValue, in := parentAttrs[attribute]; in {
				self.attributes[attribute] = parentValue
			} else {
				delete(self.attributes, attribute)
			}
		}
	}
	// Manage text by creating children
	if self.tag == "text" || self.tag == "textPath" || self.tag == "a" {
		self.children, _ = self.textChildren(element, true, true)
	}

	if parentChildren {
		self.children = make([]*Node, len(parent.children))
		for i, child := range parent.children {
			self.children[i] = NewNode(child.element, style, self.urlFetcher, self, false, nil)
		}
	} else if len(self.children) == 0 {
		for _, child := range nodeChildren(element, true) {
			if matchFeatures(child) {
				self.children = append(self.children, NewNode(child, style, self.urlFetcher, self, false, nil))
				if self.tag == "switch" {
					break
				}
			}
		}
	}
	return self
}

func (self Node) fetchUrl(u *url.URL, resourceType string) ([]byte, error) {
	return readUrl(u, self.urlFetcher, resourceType)
}

// Create children and return them. textRoot=false
func (self *Node) textChildren(element *html.Node, trailingSpace, textRoot bool) ([]*Node, bool) {
	var children []*Node
	space := "{http://www.w3.org/XML/1998/namespace}space"
	preserve := self.attributes[space] == "preserve"
	self.text = handleWhiteSpaces(extractText(element), preserve)
	if trailingSpace && !preserve {
		self.text = strings.TrimLeft(self.text, " ")
	}
	originalRotate := rotations(self)
	rotate := originalRotate
	if len(originalRotate) != 0 {
		popRotation(self, originalRotate, rotate)
	}
	if self.text != "" {
		trailingSpace = strings.HasSuffix(self.text, " ")
	}
	for _, child := range nodeChildren(element, true) {
		if child.Type == html.TextNode {
			continue
		}
		var childNode *Node
		if child.Data == "http://www.w3.org/2000/svg:tref" || child.Data == "tref" {
			href := getAttr(child, "{http://www.w3.org/1999/xlink}href")
			if href == "" {
				href = getAttr(child, "href")
			}
			u, err := url.Parse(href)
			if err != nil {
				log.Printf("invalid url %s", href)
			}
			childNode = NewNode(child, self.style, self.urlFetcher,
				self, true, u)
			childNode.tag = "tspan"
			// Retrieve the referenced node and get its flattened text
			// and remove the node children.
			childNode.text = flatten(child)
		} else {
			childNode = NewNode(child, self.style, self.urlFetcher, self, false, nil)
		}
		childPreserve := childNode.attributes[space] == "preserve"
		childNode.text = handleWhiteSpaces(extractText(child), childPreserve)
		childNode.children, _ = childNode.textChildren(child, trailingSpace, false)
		trailingSpace = strings.HasSuffix(childNode.text, " ")
		if _, in := childNode.attributes["rotate"]; len(originalRotate) != 0 && !in {
			popRotation(childNode, originalRotate, rotate)
		}
		children = append(children, childNode)
		if tail := extractTail(child); tail != "" {
			anonymousEtree := html.Node{Type: html.ElementNode, Data: "{http://www.w3.org/2000/svg}tspan"}
			anonymous := NewNode(&anonymousEtree, self.style, self.urlFetcher, self, false, nil)
			anonymous.text = handleWhiteSpaces(tail, preserve)
			if len(originalRotate) != 0 {
				popRotation(anonymous, originalRotate, rotate)
			}
			if trailingSpace && !preserve {
				anonymous.text = strings.TrimLeft(anonymous.text, " ")
			}
			if anonymous.text != "" {
				trailingSpace = strings.HasSuffix(anonymous.text, " ")
			}
			children = append(children, anonymous)
		}
	}

	if textRoot && len(children) == 0 && !preserve {
		self.text = strings.TrimRight(self.text, " ")
	}

	return children, trailingSpace
}

func (self Node) getHref() string {
	href := self.attributes["{http://www.w3.org/1999/xlink}href"]
	if href == "" {
		href = self.attributes["href"]
	}
	return href
}

// unfold html and body tags.
// extractSVG will panic if root is not the result of html.Parse()
func extractSVG(root *html.Node) *html.Node {
	svgNode := root.FirstChild.LastChild.FirstChild
	return svgNode
}

func Parse(input io.Reader, urlFetcher UrlFectcher, url string) *Node {

	if urlFetcher == nil {
		urlFetcher = fetch
	}

	_rootNode, err := html.Parse(input)

	root := extractSVG(_rootNode)

	style := parseStylesheets(self, url)
	self := NewNode(root, style, urlFetcher, parent, parentChildren, url)
	self.root = true
	return self
}
