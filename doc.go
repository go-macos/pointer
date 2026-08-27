// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

// Package pointer says where the mouse pointer is and puts it somewhere else.
//
// It exists because of a measured failure. An application that shows one display
// on another -- a viewer, a head-up surface, a desk of captured screens inside a
// pair of glasses -- gives a person a picture of somewhere the pointer can go and
// no way to get it there. Dragging the mouse across a display whose content is a
// capture of elsewhere means dragging it BLIND: the picture does not show where
// the pointer is, because the pointer is not on the display being captured. The
// way out, for the person who met this, was unplugging the glasses.
//
// One key that puts the pointer on the screen somebody is looking at fixes that,
// and this is what such a key needs.
//
// # No permission, and no synthetic events
//
// Moving the pointer is CGWarpMouseCursorPosition, which is not an event: it
// asks the window server to put the cursor somewhere. It needs neither
// Accessibility nor Input Monitoring, and it works in a plain unbundled Go
// binary -- unlike CGEventPost, which is silently refused to one (measured in
// go-macos/hotkey, twice, with the evidence).
//
// Nothing here presses a button or sends a keystroke, on purpose. A package that
// can move the pointer is a convenience; one that can click is a robot, and the
// two want different questions asked of them.
//
// # Coordinates
//
// Everything here is in CoreGraphics GLOBAL DISPLAY SPACE: pixels, origin at the
// TOP-LEFT of the main display, y growing DOWNWARDS, all displays in one
// arrangement. It is the space CGDisplayBounds and CGWarpMouseCursorPosition
// speak, and the space go-widgets/window reports its screens in.
//
// It is NOT AppKit's. -[NSEvent mouseLocation] is bottom-left with y growing up,
// which is the other convention and half a screen out if the two are mixed. This
// package never touches AppKit, so there is one convention here and it is stated.
package pointer
