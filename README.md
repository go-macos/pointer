# pointer

[![ci](https://github.com/go-macos/pointer/actions/workflows/ci.yml/badge.svg)](https://github.com/go-macos/pointer/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-macos/pointer.svg)](https://pkg.go.dev/github.com/go-macos/pointer)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)

Where the macOS mouse pointer is, and putting it somewhere. Pure Go through
purego, `CGO_ENABLED=0`.

```go
p, _ := pointer.Position()          // in CoreGraphics global display space
_ = pointer.MoveTo(p)               // exactly there
_ = pointer.MoveToDisplay(displayID) // the middle of one display
ids, _ := pointer.Displays()        // which displays are attached
```

## Why it exists

A measured failure. An application that shows one display on another — a viewer,
a head-up surface, a desk of captured screens inside a pair of glasses — gives a
person a picture of somewhere the pointer can go and **no way to get it there**.
Dragging the mouse across a display whose content is a capture of elsewhere means
dragging it blind: the picture does not show where the pointer is, because the
pointer is not on the display being captured.

The way out, for the person who met this, was unplugging the glasses. One key
that puts the pointer on the screen they are looking at fixes it, and this is
what such a key needs.

## What it will not do

**Press anything.** Moving the pointer is `CGWarpMouseCursorPosition`, which is
not an event: it asks the window server to put the cursor somewhere. It needs
neither Accessibility nor Input Monitoring, and it works in a plain unbundled Go
binary — unlike `CGEventPost`, which is silently refused to one (measured in
[go-macos/hotkey](https://github.com/go-macos/hotkey), twice, with the evidence).

A package that can move the pointer is a convenience; one that can click is a
robot, and the two want different questions asked of them.

## Coordinates

Everything is in CoreGraphics **global display space**: pixels, origin at the
**top-left** of the main display, y growing **downwards**, all displays in one
arrangement. It is the space `CGDisplayBounds` and `CGWarpMouseCursorPosition`
speak, and the space [go-widgets/window](https://github.com/go-widgets/window)
reports its screens in.

It is *not* AppKit's. `-[NSEvent mouseLocation]` is bottom-left with y growing
up, which is the other convention and half a screen out if the two are mixed.
This package never touches AppKit, so there is one convention here and it is
stated.

## Two details the platform imposes

- **A warp leaves the mouse briefly disconnected from the cursor.** The window
  server damps the mouse for about a quarter of a second so a physical nudge does
  not fight the jump, and a person who warps and immediately moves the mouse
  finds it stuck. `MoveTo` re-associates the two, which ends that.
- **`CGDisplayBounds` answers with the empty rectangle for a display it does not
  know** rather than with an error — and "the origin, no size" is a place a
  caller can accidentally warp to. `Bounds` turns it into `ErrNoDisplay`.

## Elsewhere

Every entry point reports `ErrUnsupported` off macOS, and that is not a gap to be
filled in silence: an X11 pointer is `XWarpPointer`, a Wayland one **cannot be
moved by a client at all**, and a Windows one is `SetCursorPos` — three different
answers with three different rules about who may ask. The types and the
coordinate space are portable and waiting.

## Licence

BSD-3-Clause.
