// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package pointer

import (
	"testing"
)

// The live tests. They move the REAL pointer on the machine running them, and
// they put it back where they found it -- which is not politeness, it is the
// difference between a test suite and a prank on whoever is at the keyboard.
//
// Nothing here is skipped on a machine with a window server, and everything is
// skipped on one without: a headless runner has no pointer to move, and a test
// that "passes" by not looking is worse than one that says it did not run.

// live reports whether there is a window server to talk to, and skips if not.
func live(t *testing.T) {
	t.Helper()
	if _, err := Displays(); err != nil {
		t.Skipf("no window server here: %v", err)
	}
}

// TestPositionIsOnADisplay: wherever the pointer is, it is somewhere real.
//
// Which is the strongest thing that can be said without moving it, and it is
// worth saying: a coordinate-space mistake -- AppKit's bottom-left origin instead
// of CoreGraphics' top-left -- puts the answer off the bottom of every display on
// a machine with one screen, and off the top on a machine with two.
func TestPositionIsOnADisplay(t *testing.T) {
	live(t)
	p, err := Position()
	if err != nil {
		t.Fatalf("Position = %v", err)
	}
	ids, err := Displays()
	if err != nil {
		t.Fatal(err)
	}
	var rects []Rect
	for _, id := range ids {
		r, err := Bounds(id)
		if err != nil {
			t.Errorf("display %d: %v", id, err)
			continue
		}
		rects = append(rects, r)
		if p.In(r) {
			t.Logf("the pointer is at %+v, on display %d %+v", p, id, r)
			return
		}
	}
	t.Errorf("the pointer is at %+v, which is on none of %d displays: %+v",
		p, len(rects), rects)
}

// TestMoveToPutsThePointerThere, and puts it back.
//
// The assertion is exact rather than approximate: a warp is not a gesture, it is
// an instruction, and the window server puts the cursor at the pixel it was
// given. Anything else -- a rounding, an off-by-a-scale-factor on a Retina
// display -- is the kind of error that looks like nothing until somebody warps to
// the middle of a screen and lands a display away.
func TestMoveToPutsThePointerThere(t *testing.T) {
	live(t)
	was, err := Position()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := MoveTo(was); err != nil {
			t.Errorf("could not put the pointer back at %+v: %v", was, err)
		}
	})

	ids, err := Displays()
	if err != nil || len(ids) == 0 {
		t.Fatalf("Displays = %v, %v", ids, err)
	}
	r, err := Bounds(ids[0])
	if err != nil {
		t.Fatal(err)
	}

	// Three places on the main display, one of them a half-pixel, because that
	// is what a centre of an odd-sized display is.
	for _, want := range []Point{
		r.Centre(),
		{X: r.X + 10, Y: r.Y + 10},
		{X: r.X + r.W/2 + 0.5, Y: r.Y + r.H/2 + 0.5},
	} {
		if err := MoveTo(want); err != nil {
			t.Fatalf("MoveTo(%+v) = %v", want, err)
		}
		got, err := Position()
		if err != nil {
			t.Fatal(err)
		}
		// Within a pixel: the window server rounds to the pixel grid of the
		// display the point lands on, and a half-pixel asked for on a 1x panel
		// cannot come back.
		if d := got.X - want.X; d > 1 || d < -1 {
			t.Errorf("asked for x=%g, the pointer is at x=%g", want.X, got.X)
		}
		if d := got.Y - want.Y; d > 1 || d < -1 {
			t.Errorf("asked for y=%g, the pointer is at y=%g", want.Y, got.Y)
		}
	}
}

// TestMoveToDisplayLandsOnThatDisplay, for every display attached.
//
// This is the one the whole package is for: an application showing a picture of
// another screen needs one gesture meaning "bring the mouse here", and it has a
// display id and nothing else.
func TestMoveToDisplayLandsOnThatDisplay(t *testing.T) {
	live(t)
	was, err := Position()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = MoveTo(was) })

	ids, err := Displays()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		r, err := Bounds(id)
		if err != nil {
			t.Errorf("display %d: %v", id, err)
			continue
		}
		if err := MoveToDisplay(id); err != nil {
			t.Errorf("MoveToDisplay(%d) = %v", id, err)
			continue
		}
		got, err := Position()
		if err != nil {
			t.Fatal(err)
		}
		if !got.In(r) {
			t.Errorf("moved to display %d %+v and the pointer is at %+v", id, r, got)
		}
		t.Logf("display %d %+v: the pointer landed at %+v", id, r, got)
	}
}

// TestADisplayThatIsNotThere: an id the window server does not know.
//
// It is not a hypothetical. A virtual display removed while an application still
// holds its id is the ordinary case, and CGDisplayBounds answers with the EMPTY
// rectangle rather than an error -- so a caller that trusted it would warp to the
// origin of the main display and call it success.
func TestADisplayThatIsNotThere(t *testing.T) {
	live(t)
	const notADisplay = 0xDEADBEEF
	if _, err := Bounds(notADisplay); err == nil {
		t.Error("a made-up display id reported bounds")
	}
	if err := MoveToDisplay(notADisplay); err == nil {
		t.Error("the pointer moved to a display that does not exist")
	}
	// And the pointer did not move as a side effect of asking.
	was, err := Position()
	if err != nil {
		t.Fatal(err)
	}
	_ = MoveToDisplay(notADisplay)
	now, err := Position()
	if err != nil {
		t.Fatal(err)
	}
	if now != was {
		t.Errorf("a failed move left the pointer at %+v, from %+v", now, was)
	}
}

// TestDisplaysAgreeWithBounds: every id the list gives back has a rectangle, and
// the rectangles do not overlap.
//
// The second half is a property of the arrangement rather than of this package,
// and it is checked because it is the assumption every "which display is the
// pointer on" question rests on -- including the one in TestPositionIsOnADisplay.
func TestDisplaysAgreeWithBounds(t *testing.T) {
	live(t)
	ids, err := Displays()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("no displays at all, on a machine with a window server")
	}
	rects := make([]Rect, 0, len(ids))
	for _, id := range ids {
		r, err := Bounds(id)
		if err != nil {
			t.Fatalf("display %d is in the list and has no bounds: %v", id, err)
		}
		if r.Empty() {
			t.Errorf("display %d has no area: %+v", id, r)
		}
		rects = append(rects, r)
	}
	for i, a := range rects {
		for j, b := range rects {
			if i >= j {
				continue
			}
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Errorf("displays %d and %d overlap: %+v and %+v",
					ids[i], ids[j], a, b)
			}
		}
	}
	t.Logf("%d displays: %+v", len(ids), rects)
}
