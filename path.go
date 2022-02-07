package gosvg

import (
	"golang.org/x/image/math/fixed"
)

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
	args [3]fixed.Point26_6 // depending on op
}

func floatToFixed(f Fl) fixed.Int26_6 {
	return fixed.Int26_6(f * 64)
}

func newMoveTo(point fixed.Point26_6) segment {
	return segment{
		op:   moveTo,
		args: [3]fixed.Point26_6{point},
	}
}

func newLineTo(point fixed.Point26_6) segment {
	return segment{
		op:   lineTo,
		args: [3]fixed.Point26_6{point},
	}
}

func newQuadTo(p1, p2 fixed.Point26_6) segment {
	return segment{
		op:   quadTo,
		args: [3]fixed.Point26_6{p1, p2},
	}
}

func newCubeTo(p1, p2, p3 fixed.Point26_6) segment {
	return segment{
		op:   cubeTo,
		args: [3]fixed.Point26_6{p1, p2, p3},
	}
}
