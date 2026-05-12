package capture

/*
#cgo linux pkg-config: x11
#include <X11/Xlib.h>
#include <stdlib.h>

static int x11_init_threads(void) {
	return XInitThreads();
}

static Window x11_focused_window(Display *dpy) {
	Window w;
	int revert;
	XGetInputFocus(dpy, &w, &revert);
	return w;
}

// Drain one pending event from the grab connection.
// If it is a key event, forward it to the target (GLFW) window so that
// Ebiten still sees it.  Returns 0 when no event was pending.
static int x11_pump_one(Display *dpy, Window target) {
	if (XPending(dpy) == 0)
		return 0;
	XEvent ev;
	XNextEvent(dpy, &ev);
	if (ev.type == KeyPress || ev.type == KeyRelease) {
		ev.xkey.window = target;
		XSendEvent(dpy, target, False, KeyPressMask | KeyReleaseMask, &ev);
		XFlush(dpy);
	}
	return 1;
}
*/
import "C"

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

func init() { C.x11_init_threads() }

type x11Grabber struct {
	mu       sync.Mutex
	grabbed  atomic.Bool
	display  *C.Display
	target   C.Window
	done     chan struct{}
}

func New() Grabber {
	return &x11Grabber{}
}

func (g *x11Grabber) Grab() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.grabbed.Load() {
		return nil
	}

	dpy := C.XOpenDisplay((*C.char)(unsafe.Pointer(nil)))
	if dpy == nil {
		return fmt.Errorf("capture: cannot open X11 display (Wayland is not supported)")
	}

	target := C.x11_focused_window(dpy)
	if target == C.None {
		C.XCloseDisplay(dpy)
		return fmt.Errorf("capture: no focused X11 window")
	}

	rc := C.XGrabKeyboard(dpy, target, C.True,
		C.GrabModeAsync, C.GrabModeAsync, C.CurrentTime)
	if rc != C.GrabSuccess {
		C.XCloseDisplay(dpy)
		return fmt.Errorf("capture: XGrabKeyboard failed (code %d)", int(rc))
	}

	C.XFlush(dpy)
	g.display = dpy
	g.target = target
	g.done = make(chan struct{})
	g.grabbed.Store(true)

	go g.pump()
	return nil
}

// pump drains events from the grab connection and forwards key events
// back to the GLFW window so the normal Ebiten input pipeline keeps working.
func (g *x11Grabber) pump() {
	defer close(g.done)
	for g.grabbed.Load() {
		if C.x11_pump_one(g.display, g.target) == 0 {
			time.Sleep(time.Millisecond)
		}
	}
}

func (g *x11Grabber) Release() error {
	if !g.grabbed.Load() {
		return nil
	}
	g.grabbed.Store(false)

	if g.done != nil {
		<-g.done
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.display != nil {
		C.XUngrabKeyboard(g.display, C.CurrentTime)
		C.XFlush(g.display)
		C.XCloseDisplay(g.display)
		g.display = nil
	}
	return nil
}

func (g *x11Grabber) IsGrabbed() bool {
	return g.grabbed.Load()
}

func (g *x11Grabber) PlatformNote() string {
	return "Total Capture requires X11. Wayland is not supported."
}
