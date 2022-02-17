package gosvg

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var (
	_ backend.Canvas       = (*Canvas)(nil)
	_ backend.GraphicState = (*state)(nil)
)

type Fl = backend.Fl

func floatToFixed(f Fl) fixed.Int26_6 {
	return fixed.Int26_6(f * 64)
}

// either a color.Color or rasterx.ColorFunc
type paintColor interface {
	toRasterxColor() interface{}
}

type plainColor parser.RGBA

type funcColor rasterx.ColorFunc

func (c plainColor) toRasterxColor() interface{} {
	return parser.RGBA(c)
}

func (c funcColor) toRasterxColor() interface{} {
	return rasterx.ColorFunc(c)
}

// argument for the rasterx.SetStroke function
type strokeOptions struct {
	lineCap rasterx.CapFunc
	// lineGap     rasterx.GapFunc // not supported by webrender
	miterLimit  fixed.Int26_6
	strokeWidth fixed.Int26_6
	lineJoin    rasterx.JoinMode
}

var (
	joinToFunc = [...]rasterx.JoinMode{
		backend.Round: rasterx.Round,
		backend.Bevel: rasterx.Bevel,
		backend.Miter: rasterx.Miter,
	}

	capToFunc = [...]rasterx.CapFunc{
		backend.ButtCap:   rasterx.ButtCap,
		backend.SquareCap: rasterx.SquareCap,
		backend.RoundCap:  rasterx.RoundCap,
	}
)

// graphic state
// to correctly handle opacity mask, each state must
// have its own target image, which is merged into the main output
// when closing the state.
type state struct {
	stroker *rasterx.Dasher
	filler  *rasterx.Filler

	strokeOptions strokeOptions
	dashes        []fixed.Int26_6
	dashOffset    fixed.Int26_6

	mat         matrix.Transform
	strokeColor paintColor
	fillColor   paintColor

	image *image.RGBA // shared output of `stroker` and `filler`

	mask *image.Alpha // optional
}

// return a new graphic state writing to `dst`
// if parent is not nil, initialise state from it
func newState(dst *image.RGBA, parent *state) state {
	r := dst.Bounds()
	dx, dy := r.Dx(), r.Dy()
	var out state
	if parent != nil {
		out = *parent
	} else {
		out.mat = matrix.Identity()
	}

	out.stroker = rasterx.NewDasher(dx, dy, rasterx.NewScannerGV(dx, dy, dst, r))
	out.filler = rasterx.NewFiller(dx, dy, rasterx.NewScannerGV(dx, dy, dst, r))
	out.image = dst
	return out
}

// SetAlphaMask inteprets `mask` as an alpha mask
func (st *state) SetAlphaMask(mask backend.Canvas) {
	st.mask = rgbToAlpha(mask.(*Canvas).state.image)
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
func (st *state) Clip(evenOdd bool) {
	panic("not implemented") // TODO: Implement
}

// Sets the color which will be used for any subsequent drawing operation.
//
// The color and alpha components are
// floating point numbers in the range 0 to 1.
// If the values passed in are outside that range, they will be clamped.
// `stroke` controls whether stroking or filling operations are concerned.
func (st *state) SetColorRgba(c parser.RGBA, stroke bool) {
	if stroke {
		st.strokeColor = plainColor(c)
	} else {
		st.fillColor = plainColor(c)
	}
}

func toPngBytes(m image.Image) ([]byte, error) {
	var b bytes.Buffer

	// Write the image into the buffer
	err := png.Encode(&b, m)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func saveToPngFile(filePath string, m image.Image) error {
	b, err := toPngBytes(m)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, b, os.ModePerm)
	return err
}

// SetColorPattern set the current paint color to the given pattern.
// A pattern acts as a fill or stroke color, but permits complex textures.
// It consists of a rectangle, fill with arbitrary content, which will be replicated
// at fixed horizontal and vertical intervals to fill an area.
// (contentWidth, contentHeight) define the size of the pattern content.
// `mat` maps the pattern’s internal coordinate system to the one
// in which it will painted.
// `stroke` controls whether stroking or filling operations are concerned.
func (st *state) SetColorPattern(pattern backend.Canvas, contentWidth backend.Fl, contentHeight backend.Fl, mat matrix.Transform, stroke bool) {
	// FIXME:
	fmt.Println(contentWidth, contentHeight, mat)
	patternPixels := pattern.(*Canvas).state.image
	var cf rasterx.ColorFunc = func(x, y int) color.Color {
		fmt.Println(x, y)
		return patternPixels.At(x, y)
	}

	if stroke {
		st.strokeColor = funcColor(cf)
	} else {
		st.fillColor = funcColor(cf)
	}
	// FIXME:
	// saveToPngFile(fmt.Sprintf("image%d.png", count), st.image)
	// saveToPngFile(fmt.Sprintf("pattern%d.png", count), pattern.(*Canvas).state.image)
	// applyPattern(st.image, pattern.(*Canvas).state.image)
	// saveToPngFile(fmt.Sprintf("after_pattern%d.png", count), st.image)
}

// SetBlendingMode sets the blending mode, which is a CSS blend mode keyword.
func (st *state) SetBlendingMode(mode string) {
	log.Println("blend mode not supported")
}

// Sets the current line width to be used by `Stroke`.
// The line width value specifies the diameter of a pen
// that is circular in user space,
// (though device-space pen may be an ellipse in general
// due to scaling / shear / rotation of the CTM).
func (st *state) SetLineWidth(width backend.Fl) {
	// TODO: correct the scaling
	a, b, c, d := st.mat.A, st.mat.B, st.mat.C, st.mat.D
	normA := math.Sqrt(float64(a*a + b*b + c*c + d*d))

	st.strokeOptions.strokeWidth = floatToFixed(Fl(normA) * width)
	st.applyStrokeOptions()
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
func (st *state) SetDash(dashes []backend.Fl, offset backend.Fl) {
	st.dashes = make([]fixed.Int26_6, len(dashes))
	for i, d := range dashes {
		st.dashes[i] = floatToFixed(d)
	}
	st.dashOffset = floatToFixed(offset)

	st.applyDashes()
}

// SetStrokeOptions sets additionnal options to be used when stroking
// (in addition to SetLineWidth and SetDash)
func (st *state) SetStrokeOptions(opts backend.StrokeOptions) {
	st.strokeOptions.miterLimit = floatToFixed(opts.MiterLimit)
	st.strokeOptions.lineCap = capToFunc[opts.LineCap]
	st.strokeOptions.lineJoin = joinToFunc[opts.LineJoin]
	st.applyStrokeOptions()
}

// GetTransform returns the current transformation matrix (CTM).
func (st *state) GetTransform() matrix.Transform {
	return st.mat
}

// Modifies the current transformation matrix (CTM)
// by applying `mt` as an additional transformation.
// The new transformation of user space takes place
// after any existing transformation.
func (st *state) Transform(mt matrix.Transform) {
	st.mat.LeftMultBy(mt)
}

// SetTextPaint adjusts how text shapes are rendered.
func (st *state) SetTextPaint(op backend.PaintOp) {
	log.Println("TextPaint is ignored")
}

// apply the current stroke params to the rasterx stroker
func (st *state) applyStrokeOptions() {
	opts := st.strokeOptions
	st.stroker.Stroker.SetStroke(opts.strokeWidth, opts.miterLimit, opts.lineCap, opts.lineCap, rasterx.RoundGap, opts.lineJoin)
}

// apply the current dash params to the rasterx stroker
func (st *state) applyDashes() {
	st.stroker.Dashes = st.dashes
	st.stroker.DashOffset = st.dashOffset
}

// TODO: handle patterns
func (st *state) applyFillColor() {
	st.filler.SetColor(st.fillColor.toRasterxColor())
}

// TODO: handle patterns
func (st *state) applyStrokeColor() {
	st.stroker.SetColor(st.strokeColor.toRasterxColor())
}

type Canvas struct {
	state  state   // current state
	states []state // stack

	rectangle [4]Fl // left, top, right, bottom

	hasPath bool // to avoid useless call to the rasterizer
}

func newCanvas(x, y, width, height Fl, dst *image.RGBA, parentState *state) *Canvas {
	return &Canvas{
		state:     newState(dst, parentState),
		rectangle: [4]Fl{x, y, x + width, y + height},
	}
}

func (cv *Canvas) State() backend.GraphicState {
	return &cv.state
}

// apply the current transformation matrix to (x, y)
func (cv *Canvas) transformPoint(x, y Fl) fixed.Point26_6 {
	x, y = cv.state.mat.Apply(x, y)
	return fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(y)}
}

func (cv *Canvas) MoveTo(x, y Fl) {
	p := cv.transformPoint(x, y)
	cv.state.stroker.Start(p)
	cv.state.filler.Start(p)
	cv.hasPath = true
}

func (cv *Canvas) LineTo(x, y Fl) {
	p := cv.transformPoint(x, y)
	cv.state.stroker.Line(p)
	cv.state.filler.Line(p)
	cv.hasPath = true
}

func (cv *Canvas) CubicTo(x1, y1, x2, y2, x3, y3 Fl) {
	p1 := cv.transformPoint(x1, y1)
	p2 := cv.transformPoint(x2, y2)
	p3 := cv.transformPoint(x3, y3)
	cv.state.stroker.CubeBezier(p1, p2, p3)
	cv.state.filler.CubeBezier(p1, p2, p3)
	cv.hasPath = true
}

func (cv *Canvas) ClosePath() {
	cv.state.stroker.Stop(true)
	cv.state.filler.Stop(true)
}

// Returns the current canvas rectangle
func (cv *Canvas) GetRectangle() (left, top, right, bottom backend.Fl) {
	return cv.rectangle[0], cv.rectangle[1], cv.rectangle[2], cv.rectangle[3]
}

// OnNewStack save the current graphic stack,
// execute the given closure, and restore the stack.
func (cv *Canvas) OnNewStack(f func()) {
	cv.states = append(cv.states, cv.state) // save
	cv.state = newState(image.NewRGBA(cv.state.image.Rect), &cv.state)

	f() // execute

	if cv.state.mask != nil {
		applyOpacityMask(cv.state.image, cv.state.mask)
	}
	L := len(cv.states)
	parent := cv.states[L-1]
	// merge the state image with its parent
	drawTo(parent.image, cv.state.image)

	// restore
	cv.state = parent
	cv.states = cv.states[:L-1]
}

// NewGroup creates a new drawing target with the given
// bounding box. It may be filled by graphic operations
// before being passed to the `DrawWithOpacity`, `SetColorPattern`
// and `DrawAsMask` methods.
func (cv *Canvas) NewGroup(x backend.Fl, y backend.Fl, width backend.Fl, height backend.Fl) backend.Canvas {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	out := newCanvas(x, y, width, height, img, &cv.state)
	return out
}

// DrawWithOpacity draw the given target to the main target, applying the given opacity (in [0,1]).
func (cv *Canvas) DrawWithOpacity(opacity backend.Fl, group backend.Canvas) {
	gr := group.(*Canvas)
	applyOpacity(gr.state.image, opacity)
	drawTo(cv.state.image, gr.state.image)
}

// Paint actually shows the current path on the target,
// either stroking, filling or doing both, according to `op`.
// The result of the operation depends on the current fill and
// stroke settings.
// After this call, the current path will be cleared.
func (cv *Canvas) Paint(op backend.PaintOp) {
	doStroke := op&backend.Stroke != 0
	doFill := op&(backend.FillEvenOdd|backend.FillNonZero) != 0

	if doStroke && cv.hasPath {
		cv.state.applyStrokeColor()
		cv.state.stroker.Stroker.Draw()
		cv.state.stroker.Clear()
	}

	if doFill && cv.hasPath {
		cv.state.applyFillColor()
		cv.state.filler.SetWinding(op&backend.FillNonZero != 0)
		cv.state.filler.Draw()
		cv.state.filler.Clear()
	}

	// reset the path and the mask
	cv.hasPath = false
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

// DrawText draws the given text using the current fill color.
// The rendering may be altered by a preivous `SetTextPaint` call.
// The fonts of the runs have been registred with `AddFont`.
func (cv *Canvas) DrawText(texts []backend.TextDrawing) {
	panic("not implemented") // TODO: Implement
}

// DrawRasterImage draws the given image at the current point, with the given dimensions.
// Typical format for image.Content are PNG, JPEG, GIF.
func (cv *Canvas) DrawRasterImage(image backend.RasterImage, width backend.Fl, height backend.Fl) {
	log.Println("nested image are not supported")
}

func toRasterxGradient(grad backend.GradientLayout, width, heigth Fl) rasterx.Gradient {
	var (
		points   [5]float64
		isRadial bool
	)
	switch grad.GradientKind.Kind {
	case "linear":
		points[0], points[1], points[2], points[3] = float64(grad.Coords[0]), float64(grad.Coords[1]), float64(grad.Coords[2]), float64(grad.Coords[3])
		isRadial = false
	case "radial":
		fx, fy, _, cx, cy, cr := float64(grad.Coords[0]), float64(grad.Coords[1]), float64(grad.Coords[2]), float64(grad.Coords[3]), float64(grad.Coords[4]), float64(grad.Coords[5])
		// in rasterx, points is cx, cy, fx, fy, r = fr; and fr is ignored
		points[0], points[1], points[2], points[3], points[4] = cx, cy, fx, fy, cr
		isRadial = true
	}
	stops := make([]rasterx.GradStop, len(grad.Positions))
	for i, offset := range grad.Positions {
		stops[i] = rasterx.GradStop{
			StopColor: grad.Colors[i],
			Offset:    float64(offset),
			Opacity:   float64(grad.Colors[i].A),
		}
	}

	mat := rasterx.Identity
	mat.D = float64(grad.ScaleY)

	return rasterx.Gradient{
		Points: points,
		Stops:  stops,
		Bounds: struct {
			X float64
			Y float64
			W float64
			H float64
		}{
			W: float64(width), H: float64(heigth),
		},
		Matrix:   mat,
		Spread:   rasterx.PadSpread,
		Units:    rasterx.UserSpaceOnUse,
		IsRadial: isRadial,
	}
}

var count = 0

// DrawGradient draws the given gradient at the current point.
// Solid gradient are already handled, meaning that only linear and radial
// must be taken care of.
func (cv *Canvas) DrawGradient(gradient backend.GradientLayout, width backend.Fl, height backend.Fl) {
	fmt.Println("DrawGradient", width, height)
	rg := toRasterxGradient(gradient, width, height)

	fn := rg.GetColorFunction(1).(rasterx.ColorFunc)

	cv.state.filler.Scanner.SetColor(fn)
	cv.Rectangle(0, 0, width, height)
	cv.state.filler.Draw()
	cv.state.filler.Clear()

	cv.hasPath = false

	saveToPngFile(fmt.Sprintf("gradient%d.png", count), cv.state.image)
	count++
}
