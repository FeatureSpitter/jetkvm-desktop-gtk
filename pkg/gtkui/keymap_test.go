package gtkui

import (
	"testing"

	"github.com/FeatureSpitter/jetkvm-desktop-gtk/pkg/input"
)

func TestGDKKeyToInputKey(t *testing.T) {
	tests := []struct {
		keyval uint
		want   input.Key
	}{
		{0x061, input.KeyA},   // 'a'
		{0x07a, input.KeyZ},   // 'z'
		{0x041, input.KeyA},   // 'A'
		{0x05a, input.KeyZ},   // 'Z'
		{0x031, input.Key1},   // '1'
		{0x030, input.Key0},   // '0'
		{0xff0d, input.KeyEnter},
		{0xff1b, input.KeyEscape},
		{0xff08, input.KeyBackspace},
		{0xff09, input.KeyTab},
		{0x020, input.KeySpace},
		{0xffbe, input.KeyF1},
		{0xffc9, input.KeyF12},
		{0xff53, input.KeyRight},
		{0xff51, input.KeyLeft},
		{0xff54, input.KeyDown},
		{0xff52, input.KeyUp},
		{0xffe3, input.KeyControlLeft},
		{0xffe1, input.KeyShiftLeft},
		{0xffe9, input.KeyAltLeft},
		{0xffeb, input.KeyMetaLeft},
	}
	for _, tt := range tests {
		got, ok := gdkKeyToInputKey(tt.keyval)
		if !ok {
			t.Errorf("gdkKeyToInputKey(0x%x) returned ok=false, want %v", tt.keyval, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("gdkKeyToInputKey(0x%x) = %v, want %v", tt.keyval, got, tt.want)
		}
	}
}

func TestGDKKeyToInputKey_Unknown(t *testing.T) {
	_, ok := gdkKeyToInputKey(0xdead)
	if ok {
		t.Error("expected ok=false for unknown keyval 0xdead")
	}
}

func TestGtkKeycodeToHID_FullRangeNamed(t *testing.T) {
	tests := []struct {
		keycode uint
		wantHID byte
		label   string
	}{
		// Row 0: Escape
		{9, 41, "Escape"},
		// Number row
		{10, 30, "1"}, {11, 31, "2"}, {12, 32, "3"}, {13, 33, "4"}, {14, 34, "5"},
		{15, 35, "6"}, {16, 36, "7"}, {17, 37, "8"}, {18, 38, "9"}, {19, 39, "0"},
		{20, 45, "Minus"}, {21, 46, "Equal"},
		{22, 42, "Backspace"}, {23, 43, "Tab"},
		// QWERTY row
		{24, 20, "Q"}, {25, 26, "W"}, {26, 8, "E"}, {27, 21, "R"}, {28, 23, "T"},
		{29, 28, "Y"}, {30, 24, "U"}, {31, 12, "I"}, {32, 18, "O"}, {33, 19, "P"},
		{34, 47, "LeftBracket"}, {35, 48, "RightBracket"},
		{36, 40, "Enter"}, {37, 224, "ControlLeft"},
		// Home row
		{38, 4, "A"}, {39, 22, "S"}, {40, 7, "D"}, {41, 9, "F"}, {42, 10, "G"},
		{43, 11, "H"}, {44, 13, "J"}, {45, 14, "K"}, {46, 15, "L"},
		{47, 51, "Semicolon"}, {48, 52, "Apostrophe"}, {49, 53, "GraveAccent"},
		{50, 225, "ShiftLeft"}, {51, 49, "Backslash"},
		// Bottom row
		{52, 29, "Z"}, {53, 27, "X"}, {54, 6, "C"}, {55, 25, "V"}, {56, 5, "B"},
		{57, 17, "N"}, {58, 16, "M"},
		{59, 54, "Comma"}, {60, 55, "Period"}, {61, 56, "Slash"},
		{62, 229, "ShiftRight"}, {63, 85, "NumpadMultiply"},
		{64, 226, "AltLeft"}, {65, 44, "Space"}, {66, 57, "CapsLock"},
		// F-keys
		{67, 58, "F1"}, {68, 59, "F2"}, {69, 60, "F3"}, {70, 61, "F4"},
		{71, 62, "F5"}, {72, 63, "F6"}, {73, 64, "F7"}, {74, 65, "F8"},
		{75, 66, "F9"}, {76, 67, "F10"},
		{77, 83, "NumLock"}, {78, 71, "ScrollLock"},
		// Numpad cluster
		{79, 95, "Numpad7"}, {80, 96, "Numpad8"}, {81, 97, "Numpad9"},
		{82, 86, "NumpadSubtract"},
		{83, 92, "Numpad4"}, {84, 93, "Numpad5"}, {85, 94, "Numpad6"},
		{86, 87, "NumpadAdd"},
		{87, 89, "Numpad1"}, {88, 90, "Numpad2"}, {89, 91, "Numpad3"},
		{90, 98, "Numpad0"}, {91, 99, "NumpadDecimal"},
		// ISO key + F11/F12
		{94, 100, "IntlBackslash"},
		{95, 68, "F11"}, {96, 69, "F12"},
		// Extended keys
		{104, 88, "NumpadEnter"}, {105, 228, "ControlRight"},
		{106, 84, "NumpadDivide"}, {107, 70, "PrintScreen"},
		{108, 230, "AltRight"},
		{110, 74, "Home"}, {111, 82, "Up"}, {112, 75, "PageUp"},
		{113, 80, "Left"}, {114, 79, "Right"},
		{115, 77, "End"}, {116, 81, "Down"}, {117, 78, "PageDown"},
		{118, 73, "Insert"}, {119, 76, "Delete"},
		{127, 72, "Pause"},
		{133, 227, "SuperLeft"}, {134, 231, "SuperRight"}, {135, 101, "ContextMenu"},
	}
	for _, tt := range tests {
		hid, ok := gtkKeycodeToHID[tt.keycode]
		if !ok {
			t.Errorf("gtkKeycodeToHID[%d] (%s) missing", tt.keycode, tt.label)
			continue
		}
		if hid != tt.wantHID {
			t.Errorf("gtkKeycodeToHID[%d] (%s) = %d, want %d", tt.keycode, tt.label, hid, tt.wantHID)
		}
	}

	if t.Failed() {
		return
	}
	if len(tests) != len(gtkKeycodeToHID) {
		t.Errorf("test table has %d entries but map has %d — add missing entries to the test", len(tests), len(gtkKeycodeToHID))
	}
}

func TestGtkKeycodeToHID_Unknown(t *testing.T) {
	_, ok := gtkKeycodeToHID[9999]
	if ok {
		t.Error("expected missing entry for unknown keycode 9999")
	}
}

func TestKeycodePreferredOverKeyval(t *testing.T) {
	// On a PT-PT keyboard, the physical key at evdev keycode 20 produces
	// GDK keyval 0x027 (apostrophe), but its physical position is HID 45
	// (Minus on US layout). The keycode path must win.
	keycodeHID, kcOK := gtkKeycodeToHID[20]
	if !kcOK {
		t.Fatal("keycode 20 not in gtkKeycodeToHID")
	}
	if keycodeHID != 45 {
		t.Fatalf("keycode 20 → HID %d, want 45 (Minus)", keycodeHID)
	}

	kvKey, kvOK := gdkKeyToInputKey(0x027) // apostrophe keyval
	if !kvOK {
		t.Fatal("keyval 0x027 not mapped by gdkKeyToInputKey")
	}
	kvHID, _ := input.KeyToHID(kvKey) // KeyApostrophe → HID 52
	if kvHID == keycodeHID {
		t.Fatal("keyval and keycode resolved to the same HID — test is meaningless")
	}
}
