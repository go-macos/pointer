// Copyright (c) the go-macos authors. All rights reserved.
//
// SPDX-License-Identifier: BSD-3-Clause

//go:build !darwin

package pointer

import (
	"errors"
	"testing"
)

// TestEveryEntryPointRefusesOffDarwin, and says the same thing.
//
// One error rather than five, and it is the same error a caller can test with
// errors.Is: an application that runs on three platforms wants to ask "can I move
// the pointer here" once, not per function.
func TestEveryEntryPointRefusesOffDarwin(t *testing.T) {
	if _, err := Position(); !errors.Is(err, ErrUnsupported) {
		t.Errorf("Position = %v", err)
	}
	if err := MoveTo(Point{X: 10, Y: 10}); !errors.Is(err, ErrUnsupported) {
		t.Errorf("MoveTo = %v", err)
	}
	if _, err := Bounds(1); !errors.Is(err, ErrUnsupported) {
		t.Errorf("Bounds = %v", err)
	}
	if err := MoveToDisplay(1); !errors.Is(err, ErrUnsupported) {
		t.Errorf("MoveToDisplay = %v", err)
	}
	if ids, err := Displays(); !errors.Is(err, ErrUnsupported) || ids != nil {
		t.Errorf("Displays = %v, %v", ids, err)
	}
}
