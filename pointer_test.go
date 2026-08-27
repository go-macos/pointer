// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

package pointer

import "testing"

// TestPointIn: the rectangle a point is in, at every edge.
//
// The edges are the whole of it. Global display space puts displays edge to edge,
// so a point on a boundary belongs to exactly one of two displays -- and a
// half-open rectangle is what makes that true rather than a coin toss.
func TestPointIn(t *testing.T) {
	r := Rect{X: 100, Y: 50, W: 200, H: 100}
	for _, c := range []struct {
		what string
		p    Point
		want bool
	}{
		{"the top-left corner", Point{100, 50}, true},
		{"inside", Point{200, 100}, true},
		{"the last pixel", Point{299.9, 149.9}, true},
		{"the right edge", Point{300, 100}, false},
		{"the bottom edge", Point{200, 150}, false},
		{"left of it", Point{99.9, 100}, false},
		{"above it", Point{200, 49.9}, false},
		{"the far corner", Point{300, 150}, false},
	} {
		if got := c.p.In(r); got != c.want {
			t.Errorf("%s %+v in %+v = %v, want %v", c.what, c.p, r, got, c.want)
		}
	}

	// A display to the left, sharing a boundary: a point on the seam is in one
	// of them and not both, which is what a warp to a centre never has to think
	// about and a caller hit-testing a pointer does.
	left := Rect{X: -100, Y: 50, W: 200, H: 100}
	seam := Point{X: 100, Y: 100}
	if seam.In(left) || !seam.In(r) {
		t.Errorf("the point on the seam is in left=%v right=%v", seam.In(left), seam.In(r))
	}
}

// TestRectCentre is the middle, and why it is the middle.
//
// A pointer put "on a display" goes to its centre rather than a corner, because a
// corner is shared with the next display along: half a pixel of rounding there
// and the pointer lands on the neighbour, which is the one place it must not.
func TestRectCentre(t *testing.T) {
	for _, c := range []struct {
		r    Rect
		want Point
	}{
		{Rect{0, 0, 1920, 1200}, Point{960, 600}},
		{Rect{-11520, 0, 1920, 1200}, Point{-10560, 600}},
		{Rect{100, 50, 200, 100}, Point{200, 100}},
		// An odd size: the centre is a half-pixel, and it stays one. Rounding it
		// here would be rounding in the wrong place -- CoreGraphics takes floats.
		{Rect{0, 0, 3, 3}, Point{1.5, 1.5}},
	} {
		if got := c.r.Centre(); got != c.want {
			t.Errorf("centre of %+v = %+v, want %+v", c.r, got, c.want)
		}
		if !c.r.Centre().In(c.r) {
			t.Errorf("the centre of %+v is not inside it", c.r)
		}
	}
}

// TestRectEmpty: a display that has gone away.
//
// CGDisplayBounds answers with the empty rectangle for an id it does not know,
// and "the origin, no size" is a place a caller can accidentally warp to -- so
// the wrapper turns it into an error and this is the test that says which
// rectangles count as gone.
func TestRectEmpty(t *testing.T) {
	for _, c := range []struct {
		r     Rect
		empty bool
	}{
		{Rect{}, true},
		{Rect{X: 100, Y: 100}, true},
		{Rect{W: 1920}, true},
		{Rect{H: 1200}, true},
		{Rect{W: -1920, H: 1200}, true},
		{Rect{W: 1920, H: 1200}, false},
	} {
		if got := c.r.Empty(); got != c.empty {
			t.Errorf("%+v empty = %v, want %v", c.r, got, c.empty)
		}
	}
}
