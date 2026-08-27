// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build !darwin

package pointer

// Every entry point reports [ErrUnsupported] here.
//
// Not a gap to be filled in silence: an X11 pointer is XWarpPointer, a Wayland
// one cannot be moved by a client at all, and a Windows one is SetCursorPos. Three
// different answers with three different rules about who is allowed to ask, and
// none of them "the same thing on another platform". The types and the
// coordinate space ([Point], [Rect]) are portable and waiting.

// Position reports [ErrUnsupported].
func Position() (Point, error) { return Point{}, ErrUnsupported }

// MoveTo reports [ErrUnsupported].
func MoveTo(Point) error { return ErrUnsupported }

// Bounds reports [ErrUnsupported].
func Bounds(uint32) (Rect, error) { return Rect{}, ErrUnsupported }

// MoveToDisplay reports [ErrUnsupported].
func MoveToDisplay(uint32) error { return ErrUnsupported }

// Displays reports [ErrUnsupported].
func Displays() ([]uint32, error) { return nil, ErrUnsupported }
