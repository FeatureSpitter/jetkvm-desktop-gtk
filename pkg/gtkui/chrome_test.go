package gtkui

import (
	"math"
	"testing"
)

func TestChromeMarginPctRoundtrip(t *testing.T) {
	tests := []struct {
		name              string
		marginX, marginY  int
		saveW, saveH      int
		restoreW, restoreH int
	}{
		{"bottom-right same size", 8, 900, 1920, 1080, 1920, 1080},
		{"bottom-right maximize→restore", 8, 900, 1920, 1080, 1024, 720},
		{"bottom-right small→large", 4, 350, 800, 600, 1920, 1080},
		{"top-right default", 8, 8, 1920, 1080, 1920, 1080},
		{"center-ish", 500, 300, 1920, 1080, 1920, 1080},
		{"center-ish resize", 500, 300, 1920, 1080, 1280, 720},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pctX, pctY := chromeMarginToPct(tt.marginX, tt.marginY, tt.saveW, tt.saveH)

			if pctX < 0 || pctX > 1 {
				t.Errorf("pctX=%f out of [0,1]", pctX)
			}
			if pctY < 0 || pctY > 1 {
				t.Errorf("pctY=%f out of [0,1]", pctY)
			}

			gotX, gotY := chromePctToMargin(pctX, pctY, tt.restoreW, tt.restoreH)

			// When restoring to the SAME size, margins must match exactly (within rounding).
			if tt.saveW == tt.restoreW && tt.saveH == tt.restoreH {
				if gotX != tt.marginX {
					t.Errorf("same-size roundtrip: marginX=%d, got %d", tt.marginX, gotX)
				}
				if gotY != tt.marginY {
					t.Errorf("same-size roundtrip: marginY=%d, got %d", tt.marginY, gotY)
				}
			}

			// Proportional check: the restored position as a fraction of the
			// new window must match the original fraction (within 1px tolerance
			// from int truncation).
			wantFracX := float64(tt.marginX) / float64(tt.saveW)
			gotFracX := float64(gotX) / float64(tt.restoreW)
			if math.Abs(wantFracX-gotFracX) > 1.0/float64(tt.restoreW)+1e-9 {
				t.Errorf("proportional X: want frac %f, got frac %f (margin %d in %d)",
					wantFracX, gotFracX, gotX, tt.restoreW)
			}

			wantFracY := float64(tt.marginY) / float64(tt.saveH)
			gotFracY := float64(gotY) / float64(tt.restoreH)
			if math.Abs(wantFracY-gotFracY) > 1.0/float64(tt.restoreH)+1e-9 {
				t.Errorf("proportional Y: want frac %f, got frac %f (margin %d in %d)",
					wantFracY, gotFracY, gotY, tt.restoreH)
			}
		})
	}
}

func TestChromeBottomRightStaysBottomRight(t *testing.T) {
	// User puts bar at bottom-right on 1920x1080 (marginX=8, marginY=900).
	// After resize to 1280x720, bar should still be at bottom-right
	// (proportionally: ~83% down from top).
	marginX, marginY := 8, 900
	pctX, pctY := chromeMarginToPct(marginX, marginY, 1920, 1080)

	restoredX, restoredY := chromePctToMargin(pctX, pctY, 1280, 720)

	// marginY/windowH ratio should be preserved
	origRatio := float64(marginY) / 1080.0
	newRatio := float64(restoredY) / 720.0
	if math.Abs(origRatio-newRatio) > 0.01 {
		t.Errorf("Y ratio shifted: orig=%.3f new=%.3f (restoredY=%d)",
			origRatio, newRatio, restoredY)
	}

	// X should scale proportionally too
	origRatioX := float64(marginX) / 1920.0
	newRatioX := float64(restoredX) / 1280.0
	if math.Abs(origRatioX-newRatioX) > 0.01 {
		t.Errorf("X ratio shifted: orig=%.3f new=%.3f (restoredX=%d)",
			origRatioX, newRatioX, restoredX)
	}
}

func TestChromePctToMarginClampsNegative(t *testing.T) {
	x, y := chromePctToMargin(-0.5, -0.1, 1920, 1080)
	if x < 0 || y < 0 {
		t.Errorf("negative margins not clamped: x=%d y=%d", x, y)
	}
}
