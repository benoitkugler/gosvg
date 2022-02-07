package gosvg

import (
	"fmt"
	"log"
	"math"

	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var _ backend.Canvas = (*Canvas)(nil)

type Fl = backend.Fl

// graphic state
type state struct {
	// stoke parameters
	capL        rasterx.CapFunc
	capT        rasterx.CapFunc
	gp          rasterx.GapFunc
	dashes      []float64
	dashOffset  float64
	miterLimit  fixed.Int26_6
	strokeWidth fixed.Int26_6
	jm          rasterx.JoinMode

	mat         matrix.Transform
	strokeColor parser.RGBA
	fillColor   parser.RGBA
}

type Canvas struct {
	// buffer for the current path
	// transformation have been resolved
	path []segment

	states []state // stack
	state  state   // current state

	rectangle [4]Fl // left, top, right, bottom
}

func NewCanvas(x, y, width, height Fl) *Canvas {
	return &Canvas{
		state: state{
			mat: matrix.Identity(),
		},
		rectangle: [4]Fl{x, y, x + width, y + height},
	}
}

// apply the current transformation matrix to (x, y)
func (cv *Canvas) transformPoint(x, y Fl) fixed.Point26_6 {
	x, y = cv.state.mat.Apply(x, y)
	return fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(y)}
}

func (cv *Canvas) MoveTo(x, y Fl) {
	p := cv.transformPoint(x, y)
	cv.path = append(cv.path, newMoveTo(p))
}

func (cv *Canvas) LineTo(x, y Fl) {
	p := cv.transformPoint(x, y)
	cv.path = append(cv.path, newLineTo(p))
}

func (cv *Canvas) CubicTo(x1, y1, x2, y2, x3, y3 Fl) {
	p1 := cv.transformPoint(x1, y1)
	p2 := cv.transformPoint(x2, y2)
	p3 := cv.transformPoint(x3, y3)
	cv.path = append(cv.path, newCubeTo(p1, p2, p3))
}

func (cv *Canvas) ClosePath() {
	cv.path = append(cv.path, segment{op: close})
}

// Returns the current canvas rectangle
func (cv *Canvas) GetRectangle() (left, top, right, bottom backend.Fl) {
	return cv.rectangle[0], cv.rectangle[1], cv.rectangle[2], cv.rectangle[3]
}

// OnNewStack save the current graphic stack,
// execute the given closure, and restore the stack.
func (cv *Canvas) OnNewStack(f func()) {
	cv.states = append(cv.states, cv.state) // save
	f()                                     // execute
	// restore
	L := len(cv.states)
	cv.state = cv.states[L-1]
	cv.states = cv.states[:L-1]
}

// NewGroup creates a new drawing target with the given
// bounding box. It may be filled by graphic operations
// before being passed to the `DrawWithOpacity`, `SetColorPattern`
// and `DrawAsMask` methods.
func (cv *Canvas) NewGroup(x backend.Fl, y backend.Fl, width backend.Fl, height backend.Fl) backend.Canvas {
	return NewCanvas(x, y, width, height)
}

// DrawWithOpacity draw the given target to the main target, applying the given opacity (in [0,1]).
func (cv *Canvas) DrawWithOpacity(opacity backend.Fl, group backend.Canvas) {
	gr := group.(*Canvas)
	fmt.Println(gr.path)
	log.Println("DrawWithOpacity not implemented", opacity) // TODO:
}

// DrawAsMask inteprets `mask` as an alpha mask
func (cv *Canvas) DrawAsMask(mask backend.Canvas) {
	panic("not implemented") // TODO: Implement
}

// Establishes a new clip region
// by intersecting the current clip region
// with the current path as it would be filled by `Fill`
// and according to the fill rule given in `evenOdd`.
//
// After `Clip`, the current path will be cleared (or closed).
//
// The current clip region affects all drawing operations
// by effectively masking out any changes to the surface
// that are outside the current clip region.
//
// Calling `Clip` can only make the clip region smaller,
// never larger, but you can call it in the `OnNewStack` closure argument,
// so that the original clip region is restored afterwards.
func (cv *Canvas) Clip(evenOdd bool) {
	panic("not implemented") // TODO: Implement
}

// Sets the color which will be used for any subsequent drawing operation.
//
// The color and alpha components are
// floating point numbers in the range 0 to 1.
// If the values passed in are outside that range, they will be clamped.
// `stroke` controls whether stroking or filling operations are concerned.
func (cv *Canvas) SetColorRgba(color parser.RGBA, stroke bool) {
	if stroke {
		cv.state.strokeColor = color
	} else {
		cv.state.fillColor = color
	}
}

// SetColorPattern set the current paint color to the given pattern.
// A pattern acts as a fill or stroke color, but permits complex textures.
// It consists of a rectangle, fill with arbitrary content, which will be replicated
// at fixed horizontal and vertical intervals to fill an area.
// (contentWidth, contentHeight) define the size of the pattern content.
// `mat` maps the pattern’s internal coordinate system to the one
// in which it will painted.
// `stroke` controls whether stroking or filling operations are concerned.
func (cv *Canvas) SetColorPattern(pattern backend.Canvas, contentWidth backend.Fl, contentHeight backend.Fl, mat matrix.Transform, stroke bool) {
	log.Println("SetColorPattern not implemented") // TODO: Implement
}

// SetBlendingMode sets the blending mode, which is a CSS blend mode keyword.
func (cv *Canvas) SetBlendingMode(mode string) {
	log.Println("blend mode not supported")
}

// Sets the current line width to be used by `Stroke`.
// The line width value specifies the diameter of a pen
// that is circular in user space,
// (though device-space pen may be an ellipse in general
// due to scaling / shear / rotation of the CTM).
func (cv *Canvas) SetLineWidth(width backend.Fl) {
	// TODO: correct the scaling
	a, b, c, d := cv.state.mat.A, cv.state.mat.B, cv.state.mat.C, cv.state.mat.D
	normA := math.Sqrt(float64(a*a + b*b + c*c + d*d))

	cv.state.strokeWidth = floatToFixed(Fl(normA) * width)
}

// Sets the dash pattern to be used by `Stroke`.
// A dash pattern is specified by dashes, a list of positive values.
// Each value provides the length of alternate "on" and "off"
// portions of the stroke.
// `offset` specifies a non negative offset into the pattern
// at which the stroke begins.
//
// Each "on" segment will have caps applied
// as if the segment were a separate sub-path.
// In particular, it is valid to use an "on" length of 0
// with `RoundCap` or `SquareCap`
// in order to distribute dots or squares along a path.
//
// If `dashes` is empty dashing is disabled.
// If it is of length 1 a symmetric pattern is assumed
// with alternating on and off portions of the size specified
// by the single value.
func (cv *Canvas) SetDash(dashes []backend.Fl, offset backend.Fl) {
	cv.state.dashes = make([]float64, len(dashes))
	for i, d := range dashes {
		cv.state.dashes[i] = float64(d)
	}
	cv.state.dashOffset = float64(offset)
}

// SetStrokeOptions sets additionnal options to be used when stroking
// (in addition to SetLineWidth and SetDash)
func (cv *Canvas) SetStrokeOptions(opts backend.StrokeOptions) {
	// TODO: Implement
	log.Println("not implemented", opts)
}

// Paint actually shows the current path on the target,
// either stroking, filling or doing both, according to `op`.
// The result of the operation depends on the current fill and
// stroke settings.
// After this call, the current path will be cleared.
func (cv *Canvas) Paint(op backend.PaintOp) {
	// TODO: Implement
	fmt.Println(cv.path)
	log.Println("not implemented")
	cv.path = cv.path[:0]
}

// GetTransform returns the current transformation matrix (CTM).
func (cv *Canvas) GetTransform() matrix.Transform {
	return cv.state.mat
}

// Modifies the current transformation matrix (CTM)
// by applying `mt` as an additional transformation.
// The new transformation of user space takes place
// after any existing transformation.
func (cv *Canvas) Transform(mt matrix.Transform) {
	cv.state.mat.LeftMultBy(mt)
}

// Adds a rectangle of the given size to the current path,
// at position ``(x, y)`` in user-space coordinates.
// (X,Y) coordinates are the top left corner of the rectangle.
// Note that this method may be expressed using MoveTo and LineTo,
// but may be implemented more efficiently.
func (cv *Canvas) Rectangle(x backend.Fl, y backend.Fl, width backend.Fl, height backend.Fl) {
	cv.MoveTo(x, y)
	cv.LineTo(x+width, y)
	cv.LineTo(x+width, y+height)
	cv.LineTo(x, y+height)
	cv.ClosePath()
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `font` argument,
// so caching is advised.
func (cv *Canvas) AddFont(font pango.Font, content []byte) *backend.Font {
	return &backend.Font{}
}

// SetTextPaint adjusts how text shapes are rendered.
func (cv *Canvas) SetTextPaint(op backend.PaintOp) {
	log.Println("TextPaint is ignored")
}

// DrawText draws the given text using the current fill color.
// The rendering may be altered by a preivous `SetTextPaint` call.
// The fonts of the runs have been registred with `AddFont`.
func (cv *Canvas) DrawText(texts []backend.TextDrawing) {
	panic("not implemented") // TODO: Implement
}

// DrawRasterImage draws the given image at the current point, with the given dimensions.
// Typical format for image.Content are PNG, JPEG, GIF.
func (cv *Canvas) DrawRasterImage(image backend.RasterImage, width backend.Fl, height backend.Fl) {
	log.Println("nested image not supported")
}

// DrawGradient draws the given gradient at the current point.
// Solid gradient are already handled, meaning that only linear and radial
// must be taken care of.
func (cv *Canvas) DrawGradient(gradient backend.GradientLayout, width backend.Fl, height backend.Fl) {
	log.Println("DrawGradient not implemented") // TODO: Implement
}
