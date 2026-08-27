// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package pointer

import (
	"fmt"
	"sync"

	"github.com/ebitengine/purego"
	"github.com/go-macos/objc"
)

const frameworkCoreGraphics = "/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics"

// The CoreGraphics entry points. All four are plain C functions on the window
// server, and none of them is an event: nothing here presses anything.
var (
	cgWarpMouseCursorPosition func(cgPoint) int32
	cgDisplayBounds           func(uint32) cgRect
	cgGetActiveDisplayList    func(uint32, *uint32, *uint32) int32
	cgAssociateMouseAndCursor func(int32) int32
	cgEventCreate             func(uintptr) uintptr
	cgEventGetLocation        func(uintptr) cgPoint
	cfRelease                 func(uintptr)
)

// cgPoint and cgRect are CGPoint and CGRect: two and four CGFloats, which on
// every 64-bit Mac is a float64. They are declared here rather than reused from
// another package because a struct passed BY VALUE through purego has to have
// exactly the C layout, and borrowing one is how that silently stops being true.
type cgPoint struct {
	X, Y float64
}

type cgRect struct {
	X, Y, W, H float64
}

var (
	loadOnce sync.Once
	loadErr  error
)

// dlopen is a package seam so a test can force the load-failure branch.
var dlopen = func(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func load() error {
	loadOnce.Do(func() { loadErr = doLoad() })
	return loadErr
}

func doLoad() error {
	if err := objc.Load(frameworkCoreGraphics); err != nil {
		return fmt.Errorf("pointer: %w", err)
	}
	cg, err := dlopen(frameworkCoreGraphics)
	if err != nil {
		return fmt.Errorf("pointer: CoreGraphics: %w", err)
	}
	purego.RegisterLibFunc(&cgWarpMouseCursorPosition, cg, "CGWarpMouseCursorPosition")
	purego.RegisterLibFunc(&cgDisplayBounds, cg, "CGDisplayBounds")
	purego.RegisterLibFunc(&cgGetActiveDisplayList, cg, "CGGetActiveDisplayList")
	purego.RegisterLibFunc(&cgAssociateMouseAndCursor, cg,
		"CGAssociateMouseAndMouseCursorPosition")
	purego.RegisterLibFunc(&cgEventCreate, cg, "CGEventCreate")
	purego.RegisterLibFunc(&cgEventGetLocation, cg, "CGEventGetLocation")
	purego.RegisterLibFunc(&cfRelease, cg, "CFRelease")
	return nil
}

// Position is where the pointer is now.
//
// Read from a fresh CGEvent rather than from AppKit, for two reasons that both
// matter: -[NSEvent mouseLocation] is bottom-left with y growing UP, which is the
// other convention and half a screen out if the two are mixed; and it needs an
// NSApplication, which a command-line tool asking where the mouse is has no
// business creating.
func Position() (Point, error) {
	if err := load(); err != nil {
		return Point{}, err
	}
	ev := cgEventCreate(0)
	if ev == 0 {
		// Documented to return NULL only on allocation failure, so this is the
		// branch nothing can reach on a working machine -- and the alternative to
		// checking it is reading a location out of a null pointer.
		return Point{}, fmt.Errorf("pointer: CGEventCreate returned nothing")
	}
	defer cfRelease(ev)
	p := cgEventGetLocation(ev)
	return Point{X: p.X, Y: p.Y}, nil
}

// MoveTo puts the pointer at p.
//
// It also re-associates the mouse with the cursor, which is not decoration. A
// warp leaves the two briefly disconnected -- the window server damps the mouse
// for about a quarter of a second so a physical nudge does not fight the jump --
// and a person who warps the pointer and immediately moves the mouse finds it
// stuck. Saying "these are the same thing again" ends that.
func MoveTo(p Point) error {
	if err := load(); err != nil {
		return err
	}
	if code := cgWarpMouseCursorPosition(cgPoint{X: p.X, Y: p.Y}); code != 0 {
		return fmt.Errorf("pointer: CGWarpMouseCursorPosition: CGError %d", code)
	}
	if code := cgAssociateMouseAndCursor(1); code != 0 {
		return fmt.Errorf("pointer: CGAssociateMouseAndMouseCursorPosition: CGError %d", code)
	}
	return nil
}

// Bounds is a display's rectangle in global display space.
//
// A display that is not there reports [ErrNoDisplay] rather than the empty
// rectangle CGDisplayBounds answers with, because "the origin, no size" is a
// place a caller can accidentally warp to.
func Bounds(display uint32) (Rect, error) {
	if err := load(); err != nil {
		return Rect{}, err
	}
	r := cgDisplayBounds(display)
	out := Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
	if out.Empty() {
		return Rect{}, fmt.Errorf("%w: %d", ErrNoDisplay, display)
	}
	return out, nil
}

// MoveToDisplay puts the pointer in the middle of one display.
//
// It is the whole reason this package exists: an application showing a person a
// picture of another screen needs one gesture that means "bring the mouse here",
// and the display id is what such an application has.
func MoveToDisplay(display uint32) error {
	r, err := Bounds(display)
	if err != nil {
		return err
	}
	return MoveTo(r.Centre())
}

// Displays are the attached displays, in the window server's own order.
//
// Offered because a caller that wants to know whether the id it is holding is
// still attached has otherwise to warp to it and see -- and because the answer
// says WHICH display, not merely how many, unlike the count-only call this wraps.
func Displays() ([]uint32, error) {
	if err := load(); err != nil {
		return nil, err
	}
	const max = 32
	ids := make([]uint32, max)
	var n uint32
	if code := cgGetActiveDisplayList(max, &ids[0], &n); code != 0 {
		return nil, fmt.Errorf("pointer: CGGetActiveDisplayList: CGError %d", code)
	}
	return ids[:n], nil
}
