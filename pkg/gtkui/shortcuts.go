package gtkui

import (
	"log"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/lkarlslund/jetkvm-desktop/pkg/hotkeys"
	"github.com/lkarlslund/jetkvm-desktop/pkg/input"
)

func (a *Application) setupShortcuts() {
	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		return a.handleShortcut(keyval, state)
	})
	a.window.AddController(keyCtrl)
}

func (a *Application) handleShortcut(keyval uint, state gdk.ModifierType) bool {
	if keyval == gdkEscape && a.activeOverlay != "" {
		a.closeOverlay()
		return true
	}

	if keyval == captureToggleGDKKey(a.prefs.CaptureToggleKey) {
		a.toggleCaptureKey()
		return true
	}

	if a.hkManager != nil && a.hkManager.Enabled() && a.ctrl != nil {
		var pressed []input.Key
		if state&gdk.ControlMask != 0 {
			pressed = append(pressed, input.KeyControlLeft)
		}
		if state&gdk.AltMask != 0 {
			pressed = append(pressed, input.KeyAltLeft)
		}
		if state&gdk.ShiftMask != 0 {
			pressed = append(pressed, input.KeyShiftLeft)
		}
		if key, ok := gdkKeyToInputKey(keyval); ok {
			pressed = append(pressed, key)
		}
		result := a.hkManager.Update(pressed)
		for _, action := range result.Actions {
			go func(act hotkeys.Action) {
				if err := a.ctrl.ExecuteRemoteHotkey(act); err != nil {
					log.Printf("[gtkui] hotkey %v failed: %v", act, err)
				}
			}(action)
		}
		if result.Consumed {
			return true
		}
	}

	return false
}

// captureToggleGDKKey maps the preference string to a GDK keyval.
func captureToggleGDKKey(name string) uint {
	switch name {
	case "F1":
		return gdkF1
	case "F2":
		return gdkF1 + 1
	case "F3":
		return gdkF1 + 2
	case "F4":
		return gdkF1 + 3
	case "F5":
		return gdkF1 + 4
	case "F6":
		return gdkF1 + 5
	case "F7":
		return gdkF1 + 6
	case "F8":
		return gdkF1 + 7
	case "F9":
		return gdkF1 + 8
	case "F10":
		return gdkF1 + 9
	case "F11":
		return gdkF1 + 10
	case "F12":
		return gdkF12
	case "Pause":
		return gdkPause
	case "ScrollLock":
		return gdkScrollLock
	default:
		return gdkScrollLock
	}
}
