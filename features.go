package gosvg

import (
	"github.com/benoitkugler/go-weasyprint/utils"
	"strings"

	go_locale "github.com/jmshal/go-locale"
	"golang.org/x/net/html"
)

// Helpers related to SVG conditional processing.

var (
	ROOT              = "http://www.w3.org/TR/SVG11/feature"
	LOCALE            = ""
	SUPPORTEDFEATURES = utils.NewSet()
)

func init() {
	locale, err := go_locale.DetectLocale()
	if err == nil {
		LOCALE = locale
	}
	for _, feature := range []string{
		"SVG",
		"SVG-static",
		"CoreAttribute",
		"Structure",
		"BasicStructure",
		"ConditionalProcessing",
		"Image",
		"Style",
		"ViewportAttribute",
		"Shape",
		"BasicText",
		"BasicPaintAttribute",
		"OpacityAttribute",
		"BasicGraphicsAttribute",
		"Marker",
		"Gradient",
		"Pattern",
		"Clip",
		"BasicClip",
		"Mask",
	} {
		SUPPORTEDFEATURES.Add(ROOT + "#" + feature)
	}
}

// Check whether ``features`` are supported by CairoSVG.
func hasFeatures(features string) bool {
	for _, feature := range strings.Split(strings.TrimSpace(features), " ") {
		if !SUPPORTEDFEATURES.Has(feature) {
			return false
		}
	}
	return true
}

// Check whether one of ``languages`` is part of the user locales.
func supportLanguages(languages string) bool {
	for _, language := range strings.Split(languages, ",") {
		language = strings.TrimSpace(language)
		if language != "" && strings.HasPrefix(LOCALE, language) {
			return true
		}
	}
	return false
}

// Check if the node matches the conditional processing attributes.
func matchFeatures(node *html.Node) bool {
	var features, languages string
	for _, attr := range node.Attr {
		if attr.Key == "requiredExtensions" {
			return false
		}
		if attr.Key == "requiredFeatures" {
			features = attr.Val
		}
		if attr.Key == "systemLanguage" {
			languages = attr.Val
		}
	}
	if features != "" && !hasFeatures(features) {
		return false
	}
	if languages != "" && !supportLanguages(languages) {
		return false
	}
	return true
}
