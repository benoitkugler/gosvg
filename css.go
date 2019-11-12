package gosvg

import (
	"fmt"
	"log"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"

	"github.com/benoitkugler/cascadia"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"golang.org/x/net/html"
)

// Handle CSS stylesheets.

type match struct {
	selector     cascadia.SelectorGroup
	declarations [][2]string
}

type matcher []match

type matchResult struct {
	specificity cascadia.Specificity
	pseudoType  string
	payload     [][2]string
}

func (m matcher) Match(element *html.Node) (out []matchResult) {
	for _, mat := range m {
		for _, sel := range mat.selector {
			if sel.Match(element) {
				out = append(out, matchResult{specificity: sel.Specificity(), pseudoType: sel.PseudoElement(), payload: mat.declarations})
			}
		}
	}
	return
}

// Find the stylesheets included in ``tree``.
func findStylesheets(tree *utils.HTMLNode) [][]parser.Token {
	// TODO: support contentStyleType on <svg>
	var out [][]parser.Token
	iter := utils.NewHtmlIterator(tree.AsHtmlNode())
	for iter.HasNext() {
		element := iter.Next()
		// http://www.w3.org/TR/SVG/styling.html#StyleElement
		if text := element.GetText(); element.Data == "{http://www.w3.org/2000/svg}style" &&
			(element.Get("type") == "" || element.Get("type") == "text/css") &&
			text != "" {

			// TODO: pass href for relative URLs
			// TODO: support media types
			// TODO: what if <style> has children elements?
			out = append(out, parser.ParseStylesheet2([]byte(text), true, true))
		}
	}
	return out
}

// Find the rules in a stylesheet.
func findStylesheetsRules(tree *html.Node, stylesheetRules []parser.Token, urlFetcher UrlFectcher, url string) []parser.QualifiedRule {
	var out []parser.QualifiedRule
	for _, rule := range stylesheetRules {
		switch rule := rule.(type) {
		case parser.AtRule:
			if rule.AtKeyword.Lower() == "import" && rule.Content == nil {
				// TODO: support media types in @import
				urlToken := parser.ParseOneComponentValue(*rule.Prelude)
				var v string
				switch urlToken := urlToken.(type) {
				case parser.StringToken:
					v = urlToken.Value
				case parser.URLToken:
					v = urlToken.Value
				default:
					continue
				}
				cssUrl := utils.UrlJoin(url, v, false, "@import")
				data, err := urlFetcher(cssUrl, "text/css")
				if err != nil {
					log.Printf("unable to fetch style at %s", cssUrl)
					continue
				}
				stylesheet := parser.ParseStylesheet2(data, false, false)
				out = append(out, findStylesheetsRules(tree, stylesheet, urlFetcher, cssUrl)...)
			}
			// TODO: support media types
			// if rule.lowerAtKeyword == "media" :
		case parser.QualifiedRule:
			out = append(out, rule)
		case parser.ParseError:
			log.Printf("error in css %s (%s) at %d:%d\n", rule.Kind, rule.Message, rule.Line, rule.Column)
		}
	}
	return out
}

func _parseDecl(parsed []parser.Token) ([][2]string, [][2]string) {
	var normalDeclarations, importantDeclarations [][2]string // name, value
	for _, declaration := range parsed {
		if err, ok := declaration.(parser.ParseError); ok {
			log.Printf("error in css %s (%s) at %d:%d\n", err.Kind, err.Message, err.Line, err.Column)
			continue
		}
		if declaration, ok := declaration.(parser.Declaration); ok && !strings.HasPrefix(string(declaration.Name), "-") {
			// Serializing perfectly good tokens just to re-parse them later :(
			value := strings.TrimSpace(parser.Serialize(declaration.Value))
			data := [2]string{declaration.Name.Lower(), value}
			if declaration.Important {
				importantDeclarations = append(importantDeclarations, data)
			} else {
				normalDeclarations = append(normalDeclarations, data)
			}
		}
	}
	return normalDeclarations, importantDeclarations
}

func parseDeclarations(input []parser.Token) ([][2]string, [][2]string) {
	return _parseDecl(parser.ParseDeclarationList(input, false, false))
}

func parseDeclarations2(input string) ([][2]string, [][2]string) {
	return _parseDecl(parser.ParseDeclarationList2(input, false, false))
}

// Find and parse the stylesheets in ``tree``.
// Return two matcher objects,
// for normal and !important declarations.
func parseStylesheets(tree *utils.HTMLNode, urlFetcher UrlFectcher, url string) styleMatchers {
	var style styleMatchers
	for _, stylesheet := range findStylesheets(tree) {
		for _, rule := range findStylesheetsRules(tree.AsHtmlNode(), stylesheet, urlFetcher, url) {
			normalDeclarations, importantDeclarations := parseDeclarations(*rule.Content)
			prelude := parser.Serialize(*rule.Prelude)
			selector, err := cascadia.ParseGroupWithPseudoElements(prelude)
			if err != nil {
				log.Printf("unsupported css selector %s : %s\n", prelude, err)
				continue
			}
			for _, sel := range selector {
				if sel.PseudoElement() != "" {
					err = fmt.Errorf("Unsupported pseudo-Element : %s", sel.PseudoElement())
					break
				}
			}
			if err != nil {
				log.Println(err)
				continue
			}

			if len(normalDeclarations) != 0 {
				style.normal = append(style.normal, match{selector: selector, declarations: normalDeclarations})
			}
			if len(importantDeclarations) != 0 {
				style.important = append(style.important, match{selector: selector, declarations: importantDeclarations})
			}

		}
	}
	return style
}

// // Get the declarations := range ``rule``.
// func getDeclarations(rule) {
//     if rule.type == "qualified-rule" {
//         for declaration := range tinycss2.parseDeclarationList(
//                 rule.content, skipComments=true, skipWhitespace=true) {
//                 }
//             value = "".join(part.serialize() for part := range declaration.value)
//             // TODO: filter out invalid values
//             yield declaration.lowerName, value, declaration.important
