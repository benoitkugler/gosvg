package gosvg

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Surface helpers.

var (
	UNITS = map[string]float64{
		"mm": 1 / 25.4,
		"cm": 1 / 2.54,
		"in": 1,
		"pt": 1 / 72.,
		"pc": 1 / 6.,
	}

	PAINTURL    = regexp.MustCompile(`(url\(.+\)) *(.*)`)
	PATHLETTERS = "achlmqstvzACHLMQSTVZ"
	RECT        = regexp.MustCompile(`rect\( ?(.+?) ?\)`)
)

// class PointError(Exception) {
//     """Exception raised when parsing a point fails."""
// }

// // Get the distance between two points.
// func distance(x1, y1, x2, y2) {
//     return ((x2 - x1) ** 2 + (y2 - y1) ** 2) ** 0.5
// }

// // Extract from value an uri && a color.
// //     See http://www.w3.org/TR/SVG/painting.html#SpecifyingPaint
// //
// func paint(value) {
//     if ! value {
//         return None, None
//     }
// }
//     value = value.strip()
//     match = PAINTURL.search(value)
//     if match {
//         source = parseUrl(match.group(1)).fragment
//         color = match.group(2) || None
//     } else {
//         source = None
//         color = value || None
//     }

//     return (source, color)

// // Return ``(width, height, viewbox)`` of ``node``.
// //     If ``reference`` is ``true``, we can rely on surface size to resolve
// //     percentages.
// //
// func nodeFormat(surface, node, reference=true) {
//     referenceSize = "xy" if reference else (0, 0)
//     width = size(surface, node.get("width", "100%"), referenceSize[0])
//     height = size(surface, node.get("height", "100%"), referenceSize[1])
//     viewbox = node.get("viewBox")
//     if viewbox {
//         viewbox = re.sub("[ \n\r\t,]+", " ", viewbox)
//         viewbox = tuple(float(position) for position := range viewbox.split())
//         width = width || viewbox[2]
//         height = height || viewbox[3]
//     } return width, height, viewbox
// }

var (
	re1 = regexp.MustCompile("(?<!e)-")
	re2 = regexp.MustCompile("[ \n\r\t,]+")
	re3 = regexp.MustCompile(`(\.[0-9-]+)(?=\.)`)
)

// Normalize a string corresponding to an array of various values.
func normalize(str string) string {
	str = strings.ReplaceAll(str, "E", "e")
	str = re1.ReplaceAllString(str, " -")
	str = re2.ReplaceAllString(str, " ")
	str = re3.ReplaceAllString(str, `\1 `)
	return strings.TrimSpace(str)
}

// // Return ``(x, y, trailingText)`` from ``string``.
// func point(surface, string) {
//     match = re.match("(.*?) (.*?)(?: |$)", string)
//     if match {
//         x, y = match.group(1, 2)
//         string = string[match.end():]
//         return (size(surface, x, "x"), size(surface, y, "y"), string)
//     } else {
//         raise PointError
//     }
// }

// // Return angle between x axis && point knowing given center.
// func pointAngle(cx, cy, px, py) {
//     return atan2(py - cy, px - cx)
// }

// // Manage the ratio preservation.
// func preserveRatio(surface, node, width=None, height=None) {
//     if node.tag == "marker" {
//         width = width || size(surface, node.get("markerWidth", "3"), "x")
//         height = height || size(surface, node.get("markerHeight", "3"), "y")
//         _, _, viewbox = nodeFormat(surface, node)
//         viewboxWidth, viewboxHeight = viewbox[2:]
//     } else if node.tag := range ("svg", "image", "g") {
//         nodeWidth, nodeHeight, _ = nodeFormat(surface, node)
//         width = width || nodeWidth
//         height = height || nodeHeight
//         viewboxWidth, viewboxHeight = node.imageWidth, node.imageHeight
//     }
// }
//     translateX = 0
//     translateY = 0
//     scaleX = width / viewboxWidth if viewboxWidth > 0 else 1
//     scaleY = height / viewboxHeight if viewboxHeight > 0 else 1

//     aspectRatio = node.get("preserveAspectRatio", "xMidYMid").split()
//     align = aspectRatio[0]
//     if align == "none" {
//         xPosition = "min"
//         yPosition = "min"
//     } else {
//         meetOrSlice = aspectRatio[1] if len(aspectRatio) > 1 else None
//         if meetOrSlice == "slice" {
//             scaleValue = max(scaleX, scaleY)
//         } else {
//             scaleValue = min(scaleX, scaleY)
//         } scaleX = scaleY = scaleValue
//         xPosition = align[1:4].lower()
//         yPosition = align[5:].lower()
//     }

//     if node.tag == "marker" {
//         translateX = -size(surface, node.get("refX", "0"), "x")
//         translateY = -size(surface, node.get("refY", "0"), "y")
//     } else {
//         translateX = 0
//         if xPosition == "mid" {
//             translateX = (width / scaleX - viewboxWidth) / 2
//         } else if xPosition == "max" {
//             translateX = width / scaleX - viewboxWidth
//         }
//     }

//         translateY = 0
//         if yPosition == "mid" {
//             translateY += (height / scaleY - viewboxHeight) / 2
//         } else if yPosition == "max" {
//             translateY += height / scaleY - viewboxHeight
//         }

//     return scaleX, scaleY, translateX, translateY

// // Get the clip ``(x, y, width, height)`` of the marker box.
// func clipMarkerBox(surface, node, scaleX, scaleY) {
//     width = size(surface, node.get("markerWidth", "3"), "x")
//     height = size(surface, node.get("markerHeight", "3"), "y")
//     _, _, viewbox = nodeFormat(surface, node)
//     viewboxWidth, viewboxHeight = viewbox[2:]
// }
//     align = node.get("preserveAspectRatio", "xMidYMid").split(" ")[0]
//     xPosition = "min" if align == "none" else align[1:4].lower()
//     yPosition = "min" if align == "none" else align[5:].lower()

//     clipX = viewbox[0]
//     if xPosition == "mid" {
//         clipX += (viewboxWidth - width / scaleX) / 2.
//     } else if xPosition == "max" {
//         clipX += viewboxWidth - width / scaleX
//     }

//     clipY = viewbox[1]
//     if yPosition == "mid" {
//         clipY += (viewboxHeight - height / scaleY) / 2.
//     } else if yPosition == "max" {
//         clipY += viewboxHeight - height / scaleY
//     }

//     return clipX, clipY, width / scaleX, height / scaleY

// // Return the quadratic points to create quadratic curves.
// func quadraticPoints(x1, y1, x2, y2, x3, y3) {
//     xq1 = x2 * 2 / 3 + x1 / 3
//     yq1 = y2 * 2 / 3 + y1 / 3
//     xq2 = x2 * 2 / 3 + x3 / 3
//     yq2 = y2 * 2 / 3 + y3 / 3
//     return xq1, yq1, xq2, yq2, x3, y3
// }

// // Rotate a point of an angle around the origin point.
// func rotate(x, y, angle) {
//     return x * cos(angle) - y * sin(angle), y * cos(angle) + x * sin(angle)
// }

// // Transform ``surface`` || ``gradient`` if supplied using ``string``.
// //     See http://www.w3.org/TR/SVG/coords.html#TransformAttribute
// //
// func transform(surface, string, gradient=None) {
//     if ! string {
//         return
//     }
// }
//     transformations = re.findall(r"(\w+) ?\( ?(.*?) ?\)", normalize(string))
//     matrix = cairo.Matrix()
//     for transformationType, transformation := range transformations {
//         values = [size(surface, value) for value := range transformation.split(" ")]
//         if transformationType == "matrix" {
//             matrix = cairo.Matrix(*values).multiply(matrix)
//         } else if transformationType == "rotate" {
//             angle = radians(float(values.pop(0)))
//             x, y = values || (0, 0)
//             matrix.translate(x, y)
//             matrix.rotate(angle)
//             matrix.translate(-x, -y)
//         } else if transformationType == "skewX" {
//             tangent = tan(radians(float(values[0])))
//             matrix = cairo.Matrix(1, 0, tangent, 1, 0, 0).multiply(matrix)
//         } else if transformationType == "skewY" {
//             tangent = tan(radians(float(values[0])))
//             matrix = cairo.Matrix(1, tangent, 0, 1, 0, 0).multiply(matrix)
//         } else if transformationType == "translate" {
//             if len(values) == 1 {
//                 values += (0,)
//             } matrix.translate(*values)
//         } else if transformationType == "scale" {
//             if len(values) == 1 {
//                 values = 2 * values
//             } matrix.scale(*values)
//         }
//     }

//     try {
//         matrix.invert()
//     } except cairo.Error {
//         // Matrix ! invertible, clip the surface to an empty path
//         activePath = surface.context.copyPath()
//         surface.context.newPath()
//         surface.context.clip()
//         surface.context.appendPath(activePath)
//     } else {
//         if gradient {
//             // When applied on gradient use already inverted matrix (mapping
//             // from user space to gradient space)
//             matrixNow = gradient.getMatrix()
//             gradient.setMatrix(matrixNow.multiply(matrix))
//         } else {
//             matrix.invert()
//             surface.context.transform(matrix)
//         }
//     }

// // Parse the rect value of a clip.
// func clipRect(string) {
//     match = RECT.search(normalize(string || ""))
//     return match.group(1).split(" ") if match else []
// }

// Retrieves the original rotations of a `text` or `tspan` node.
func rotations(node *Node) []float64 {
	var out []float64
	if r, in := node.attributes["rotate"]; in {
		for _, i := range strings.Split(normalize(r), " ") {
			v, err := strconv.ParseFloat(i, 64)
			if err != nil {
				log.Printf("invalid float litteral in rotate : %s \n", i)
				continue
			}
			out = append(out, v)
		}
	}
	return out
}

// Removes the rotations of a node that are already used.
func popRotation(node *Node, originalRotate, rotate []float64) {
	var rs []string
	for _ = range node.text {
		var r float64
		if len(rotate) != 0 {
			r, rotate = rotate[0], rotate[1:]
		} else {
			r = originalRotate[len(originalRotate)-1]
		}
		rs = append(rs, fmt.Sprintf("%f", r))
	}
	node.attributes["rotate"] = strings.Join(rs, " ")
}

// // Returns a list with the current letter"s positions (x, y && rotation).
// //     E.g.: for letter "L" with positions x = 10, y = 20 && rotation = 30:
// //     >>> [[10, 20, 30], "L"]
// //     Store the last value of each position && pop the first one := range order to
// //     avoid setting an x,y || rotation value that have already been used.
// //
// func zipLetters(xl, yl, dxl, dyl, rl, word) {
//     return (
//         ([pl.pop(0) if pl else None for pl := range (xl, yl, dxl, dyl, rl)], char)
//         for char := range word)
// }

// Flatten the text of a node and its children.
func flatten(node *html.Node) string {
	flattenedText := []string{extractText(node)}
	for _, child := range nodeChildren(node, true) {
		flattenedText = append(flattenedText, flatten(child))
		flattenedText = append(flattenedText, extractTail(child))
		node.RemoveChild(child)
	}
	return strings.Join(flattenedText, "")
}

// // Replace a ``string`` with units by a float value.
// //     If ``reference`` is a float, it is used as reference for percentages. If it
// //     is ``"x"``, we use the viewport width as reference. If it is ``"y"``, we
// //     use the viewport height as reference. If it is ``"xy"``, we use
// //     ``(viewportWidth ** 2 + viewportHeight ** 2) ** .5 / 2 ** .5`` as
// //     reference.
// //
// func size(surface, string, reference="xy") {
//     if ! string {
//         return 0
//     }
// }
//     try {
//         return float(string)
//     } except ValueError {
//         // Not a float, try something else
//         pass
//     }

//     // No surface (for parsing only)
//     if surface  == nil  {
//         return 0
//     }

//     string = normalize(string).split(" ", 1)[0]
//     if string.endswith("%") {
//         if reference == "x" {
//             reference = surface.contextWidth || 0
//         } else if reference == "y" {
//             reference = surface.contextHeight || 0
//         } else if reference == "xy" {
//             reference = (
//                 (surface.contextWidth ** 2 +
//                  surface.contextHeight ** 2) ** .5 /
//                 2 ** .5)
//         } return float(string[:-1]) * reference / 100
//     } else if string.endswith("em") {
//         return surface.fontSize * float(string[:-2])
//     } else if string.endswith("ex") {
//         // Assume that 1em == 2ex
//         return surface.fontSize * float(string[:-2]) / 2
//     }

//     for unit, coefficient := range UNITS.items() {
//         if string.endswith(unit) {
//             number = float(string[:-len(unit)])
//             return number * (surface.dpi * coefficient if coefficient else 1)
//         }
//     }

//     // Unknown size
//     return 0
