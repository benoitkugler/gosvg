package gosvg

import (
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
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
	node, err := Parse(strings.NewReader(s), nil, "")
	fmt.Println(node, err)
}
