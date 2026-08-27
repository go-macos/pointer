// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

package pointer

import "errors"

// Errors reported by the package. They are stable and may be tested with
// errors.Is.
var (
	// ErrUnsupported is returned by every entry point on non-darwin platforms.
	// The pointer of a window server is that window server's to move.
	ErrUnsupported = errors.New("pointer: unsupported on this platform (darwin only)")

	// ErrNoDisplay reports a display id the window server does not know, or one
	// that has gone away since the caller looked. It is an ordinary event with an
	// external panel and routine with a virtual one, which is why it is an error
	// and not a panic.
	ErrNoDisplay = errors.New("pointer: no such display")
)

// A Point is a place in CoreGraphics global display space: pixels, origin at the
// top-left of the main display, y growing downwards.
//
// Floating point because that is what CoreGraphics uses, and because a display
// with a backing factor of two has half-pixels in this space that are whole
// pixels on the panel.
type Point struct {
	X, Y float64
}

// In reports whether the point is inside the rectangle.
func (p Point) In(r Rect) bool {
	return p.X >= r.X && p.X < r.X+r.W && p.Y >= r.Y && p.Y < r.Y+r.H
}

// A Rect is a rectangle in the same space: a display's bounds, chiefly.
type Rect struct {
	X, Y, W, H float64
}

// Centre is the middle of the rectangle.
//
// It is where a pointer put "on a display" goes. The middle rather than a corner
// because a corner is shared with the next display along: half a pixel of
// rounding there and the pointer lands on the neighbour, which is the one place
// it must not.
func (r Rect) Centre() Point {
	return Point{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

// Empty reports whether the rectangle has no area. A display that has gone away
// reports one of these rather than an error, because CGDisplayBounds does.
func (r Rect) Empty() bool { return r.W <= 0 || r.H <= 0 }
