package gtkui

import "github.com/lkarlslund/jetkvm-desktop/pkg/input"

// GDK key values (from gdk/gdkkeysyms.h). Only the subset we need.
const (
	gdkA            = 0x061
	gdkZ            = 0x07a
	gdk0            = 0x030
	gdk9            = 0x039
	gdkReturn       = 0xff0d
	gdkEscape       = 0xff1b
	gdkBackSpace    = 0xff08
	gdkTab          = 0xff09
	gdkSpace        = 0x020
	gdkMinus        = 0x02d
	gdkEqual        = 0x03d
	gdkBracketLeft  = 0x05b
	gdkBracketRight = 0x05d
	gdkBackslash    = 0x05c
	gdkSemicolon    = 0x03b
	gdkApostrophe   = 0x027
	gdkGraveAccent  = 0x060
	gdkComma        = 0x02c
	gdkPeriod       = 0x02e
	gdkSlash        = 0x02f
	gdkCapsLock     = 0xffe5
	gdkF1           = 0xffbe
	gdkF12          = 0xffc9
	gdkF13          = 0xffca
	gdkF24          = 0xffd5
	gdkPrint        = 0xff61
	gdkScrollLock   = 0xff14
	gdkPause        = 0xff13
	gdkMenu         = 0xff67
	gdkInsert       = 0xff63
	gdkHome         = 0xff50
	gdkPageUp       = 0xff55
	gdkDelete       = 0xffff
	gdkEnd          = 0xff57
	gdkPageDown     = 0xff56
	gdkRight        = 0xff53
	gdkLeft         = 0xff51
	gdkDown         = 0xff54
	gdkUp           = 0xff52
	gdkNumLock      = 0xff7f
	gdkKPDivide     = 0xffaf
	gdkKPMultiply   = 0xffaa
	gdkKPSubtract   = 0xffad
	gdkKPAdd        = 0xffab
	gdkKPEnter      = 0xff8d
	gdkKP1          = 0xffb1
	gdkKP0          = 0xffb0
	gdkKPDecimal    = 0xffae
	gdkKPEqual      = 0xffbd
	gdkControlL     = 0xffe3
	gdkShiftL       = 0xffe1
	gdkAltL         = 0xffe9
	gdkSuperL       = 0xffeb
	gdkControlR     = 0xffe4
	gdkShiftR       = 0xffe2
	gdkAltR         = 0xffea
	gdkSuperR       = 0xffec
)

// gtkKeycodeToHID maps a GTK hardware keycode (evdev + 8 on Linux) directly
// to a USB HID scancode. This gives us the physical key position, which is
// what a KVM must forward — the remote OS applies its own layout.
var gtkKeycodeToHID = map[uint]byte{
	9: 41, // Escape
	10: 30, 11: 31, 12: 32, 13: 33, 14: 34, 15: 35, 16: 36, 17: 37, 18: 38, 19: 39, // 1-0
	20: 45, // Minus
	21: 46, // Equal
	22: 42, // Backspace
	23: 43, // Tab
	24: 20, 25: 26, 26: 8, 27: 21, 28: 23, 29: 28, 30: 24, 31: 12, 32: 18, 33: 19, // Q-P
	34: 47, // LeftBracket
	35: 48, // RightBracket
	36: 40, // Enter
	37: 224, // ControlLeft
	38: 4, 39: 22, 40: 7, 41: 9, 42: 10, 43: 11, 44: 13, 45: 14, 46: 15, // A-L
	47: 51, // Semicolon
	48: 52, // Apostrophe
	49: 53, // GraveAccent
	50: 225, // ShiftLeft
	51: 49, // Backslash
	52: 29, 53: 27, 54: 6, 55: 25, 56: 5, 57: 17, 58: 16, // Z-M
	59: 54, // Comma
	60: 55, // Period
	61: 56, // Slash
	62: 229, // ShiftRight
	63: 85, // NumpadMultiply
	64: 226, // AltLeft
	65: 44, // Space
	66: 57, // CapsLock
	67: 58, 68: 59, 69: 60, 70: 61, 71: 62, 72: 63, 73: 64, 74: 65, 75: 66, 76: 67, // F1-F10
	77: 83,  // NumLock
	78: 71,  // ScrollLock
	79: 95, 80: 96, 81: 97, // Numpad7-9
	82: 86, // NumpadSubtract
	83: 92, 84: 93, 85: 94, // Numpad4-6
	86: 87, // NumpadAdd
	87: 89, 88: 90, 89: 91, // Numpad1-3
	90: 98,  // Numpad0
	91: 99,  // NumpadDecimal
	94: 100, // IntlBackslash (ISO key between LShift and Z)
	95: 68,  // F11
	96: 69,  // F12
	104: 88,  // NumpadEnter
	105: 228, // ControlRight
	106: 84,  // NumpadDivide
	107: 70,  // PrintScreen
	108: 230, // AltRight
	110: 74,  // Home
	111: 82,  // Up
	112: 75,  // PageUp
	113: 80,  // Left
	114: 79,  // Right
	115: 77,  // End
	116: 81,  // Down
	117: 78,  // PageDown
	118: 73,  // Insert
	119: 76,  // Delete
	127: 72,  // Pause
	133: 227, // SuperLeft
	134: 231, // SuperRight
	135: 101, // ContextMenu
}

func gdkKeyToInputKey(keyval uint) (input.Key, bool) {
	// Lowercase letters a-z
	if keyval >= gdkA && keyval <= gdkZ {
		return input.KeyA + input.Key(keyval-gdkA), true
	}
	// Uppercase A-Z map to same keys
	if keyval >= 0x041 && keyval <= 0x05a {
		return input.KeyA + input.Key(keyval-0x041), true
	}
	// Digits 0-9
	if keyval >= gdk0 && keyval <= gdk9 {
		// input.Key order: Key1..Key9, Key0
		if keyval == gdk0 {
			return input.Key0, true
		}
		return input.Key1 + input.Key(keyval-gdk0-1), true
	}
	// F1-F12
	if keyval >= gdkF1 && keyval <= gdkF12 {
		return input.KeyF1 + input.Key(keyval-gdkF1), true
	}
	// F13-F24
	if keyval >= gdkF13 && keyval <= gdkF24 {
		return input.KeyF13 + input.Key(keyval-gdkF13), true
	}
	// Numpad 0-9
	if keyval >= gdkKP0 && keyval <= gdkKP0+9 {
		// input.Key order: KeyNumpad1..KeyNumpad9, KeyNumpad0
		if keyval == gdkKP0 {
			return input.KeyNumpad0, true
		}
		return input.KeyNumpad1 + input.Key(keyval-gdkKP1), true
	}

	switch keyval {
	case gdkReturn:
		return input.KeyEnter, true
	case gdkEscape:
		return input.KeyEscape, true
	case gdkBackSpace:
		return input.KeyBackspace, true
	case gdkTab:
		return input.KeyTab, true
	case gdkSpace:
		return input.KeySpace, true
	case gdkMinus:
		return input.KeyMinus, true
	case gdkEqual:
		return input.KeyEqual, true
	case gdkBracketLeft:
		return input.KeyLeftBracket, true
	case gdkBracketRight:
		return input.KeyRightBracket, true
	case gdkBackslash:
		return input.KeyBackslash, true
	case gdkSemicolon:
		return input.KeySemicolon, true
	case gdkApostrophe:
		return input.KeyApostrophe, true
	case gdkGraveAccent:
		return input.KeyGraveAccent, true
	case gdkComma:
		return input.KeyComma, true
	case gdkPeriod:
		return input.KeyPeriod, true
	case gdkSlash:
		return input.KeySlash, true
	case gdkCapsLock:
		return input.KeyCapsLock, true
	case gdkPrint:
		return input.KeyPrintScreen, true
	case gdkScrollLock:
		return input.KeyScrollLock, true
	case gdkPause:
		return input.KeyPause, true
	case gdkMenu:
		return input.KeyContextMenu, true
	case gdkInsert:
		return input.KeyInsert, true
	case gdkHome:
		return input.KeyHome, true
	case gdkPageUp:
		return input.KeyPageUp, true
	case gdkDelete:
		return input.KeyDelete, true
	case gdkEnd:
		return input.KeyEnd, true
	case gdkPageDown:
		return input.KeyPageDown, true
	case gdkRight:
		return input.KeyRight, true
	case gdkLeft:
		return input.KeyLeft, true
	case gdkDown:
		return input.KeyDown, true
	case gdkUp:
		return input.KeyUp, true
	case gdkNumLock:
		return input.KeyNumLock, true
	case gdkKPDivide:
		return input.KeyNumpadDivide, true
	case gdkKPMultiply:
		return input.KeyNumpadMultiply, true
	case gdkKPSubtract:
		return input.KeyNumpadSubtract, true
	case gdkKPAdd:
		return input.KeyNumpadAdd, true
	case gdkKPEnter:
		return input.KeyNumpadEnter, true
	case gdkKPDecimal:
		return input.KeyNumpadDecimal, true
	case gdkKPEqual:
		return input.KeyNumpadEqual, true
	case gdkControlL:
		return input.KeyControlLeft, true
	case gdkShiftL:
		return input.KeyShiftLeft, true
	case gdkAltL:
		return input.KeyAltLeft, true
	case gdkSuperL:
		return input.KeyMetaLeft, true
	case gdkControlR:
		return input.KeyControlRight, true
	case gdkShiftR:
		return input.KeyShiftRight, true
	case gdkAltR:
		return input.KeyAltRight, true
	case gdkSuperR:
		return input.KeyMetaRight, true
	}
	return input.KeyUnknown, false
}
