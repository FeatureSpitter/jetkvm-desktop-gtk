//go:build !linux

package gtkui

import "github.com/diamondburned/gotk4/pkg/gtk/v4"

func centerWindow(_ *gtk.ApplicationWindow) {
	// On non-Linux platforms GTK4 handles initial window placement.
}

func monitorWorkarea() (int, int) {
	return 0, 0
}
