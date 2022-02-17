package gosvg

import (
	"fmt"
	"image"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

type rasterizer interface {
	rasterx.Scanner
	rasterx.Adder
}

// element represents one graphic object that can be
// rasterized in the final output image
// elements are specified in their own abstract coordinates,
// alongside a transformation matrix (either from the state or explicitely for patterns)
// which specifies how they should actually be painted on the final image
type element interface {
	// update the element to acount for `mat`
	transform(mt matrix.Transform)
}

// group is a list of shapes, either used as primary content,
// or as alphaMask or pattern.
type group struct {
	alphaMask *group // optional
	shapes    []shape
	opacity   Fl
}

// abstractColor describe a way of stroking or filling
// a shape
type abstractColor interface {
	setOn(dst rasterx.Scanner, mat matrix.Transform)
}

func (c plainColor) setOn(dst rasterx.Scanner, _ matrix.Transform) {
	dst.SetColor(parser.RGBA(c))
}

type gradientColor backend.GradientLayout

func (gr *gradientColor) setOn(dst rasterx.Scanner, mat matrix.Transform) {
	// TODO:
}

type shape struct {
	fillColor, strokeColor abstractColor // nil for no painting
	path                   path
	nonZeroWinding         bool // when filling, use non zero winding rule over even-odd rule
}

func (s shape) rasterize(filler *rasterx.Filler, stroker *rasterx.Dasher, mt matrix.Transform) {
	if s.fillColor != nil {
		filler.SetWinding(s.nonZeroWinding)
		s.path.rasterize(filler, mt)
		s.fillColor.setOn(filler, mt)
		filler.Draw()
		filler.Clear()
	}

	if s.strokeColor != nil {
		s.path.rasterize(stroker, mt)
		s.strokeColor.setOn(stroker, mt)
		stroker.Draw()
		filler.Clear()
	}
}

type path []segment

func (p path) rasterize(dst rasterx.Adder, mt matrix.Transform) {
	for _, element := range p {
		element.rasterize(dst, mt)
	}
}

func newRectangle(x, y, width, height Fl) path {
	return path{
		newMoveTo(point{x, y}),
		newLineTo(point{x + width, y}),
		newLineTo(point{x + width, y + height}),
		newLineTo(point{x, y + height}),
		segment{op: close},
	}
}

type point struct {
	x, y Fl
}

func (p point) String() string {
	return fmt.Sprintf("(%f,%f)", p.x, p.y)
}

func (p *point) transform(mt matrix.Transform) { p.x, p.y = mt.Apply(p.x, p.y) }

func (p point) toFixed(mat matrix.Transform) fixed.Point26_6 {
	x, y := mat.Apply(p.x, p.y)
	return fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(y)}
}

type op uint8

const (
	moveTo op = iota
	lineTo
	quadTo
	cubeTo
	close
)

type segment struct {
	op   op
	args [3]point // depending on op
}

func (s segment) rasterize(dst rasterx.Adder, mt matrix.Transform) {
	switch s.op {
	case moveTo:
		dst.Start(s.args[0].toFixed(mt))
	case lineTo:
		dst.Line(s.args[0].toFixed(mt))
	case quadTo:
		dst.QuadBezier(s.args[0].toFixed(mt), s.args[1].toFixed(mt))
	case cubeTo:
		dst.CubeBezier(s.args[0].toFixed(mt), s.args[1].toFixed(mt), s.args[2].toFixed(mt))
	case close:
		dst.Stop(true)
	}
}

func newMoveTo(pt point) segment {
	return segment{
		op:   moveTo,
		args: [3]point{pt},
	}
}

func newLineTo(pt point) segment {
	return segment{
		op:   lineTo,
		args: [3]point{pt},
	}
}

func newQuadTo(p1, p2 point) segment {
	return segment{
		op:   quadTo,
		args: [3]point{p1, p2},
	}
}

func newCubeTo(p1, p2, p3 point) segment {
	return segment{
		op:   cubeTo,
		args: [3]point{p1, p2, p3},
	}
}

func (s segment) String() string {
	switch s.op {
	case moveTo:
		return fmt.Sprintf("<M %s>", s.args[0])
	case lineTo:
		return fmt.Sprintf("<L %s>", s.args[0])
	case quadTo:
		return fmt.Sprintf("<Q %s %s>", s.args[0], s.args[1])
	case cubeTo:
		return fmt.Sprintf("<C %s %s %s>", s.args[0], s.args[1], s.args[2])
	case close:
		return "<close>"
	}
	return "<invalid>"
}

func (gr group) rasterize(filler *rasterx.Filler, stroker *rasterx.Dasher, mt matrix.Transform) {
	// handle mask and opacity
	for _, sh := range gr.shapes {
		sh.rasterize(filler, stroker, mt)
	}
}

// rasterize draws `gr` contents on `dst`, using `mat` to map internal
// element `coordinates` to pixel indices
// For instance, using the identity matrix inteprets coordinates as pixel indices,
// with (0,0) between the top left pixel, and the y axis growing downwards.
func rasterize(dst *image.RGBA, mat matrix.Transform) {
	// TODO: handle opacity and mask
	r := dst.Bounds()
	dx, dy := r.Dx(), r.Dy()
	var out state
	// if parent != nil {
	// 	out = *parent
	// } else {
	// 	out.mat = matrix.Identity()
	// }

	out.stroker = rasterx.NewDasher(dx, dy, rasterx.NewScannerGV(dx, dy, dst, r))
	out.filler = rasterx.NewFiller(dx, dy, rasterx.NewScannerGV(dx, dy, dst, r))
}
