package gtkui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type chromeAction int

const (
	chromeReconnect chromeAction = iota
	chromePaste
	chromeMedia
	chromeSerial
	chromeWoL
	chromeCapture
	chromeStats
	chromeFullscreen
	chromeSettings
)

type chromeButton struct {
	icon    string
	tooltip string
	action  chromeAction
}

var chromeButtons = []chromeButton{
	{"view-refresh-symbolic", "Reconnect", chromeReconnect},
	{"edit-paste-symbolic", "Paste text", chromePaste},
	{"media-optical-symbolic", "Virtual media", chromeMedia},
	{"utilities-terminal-symbolic", "Serial console", chromeSerial},
	{"system-shutdown-symbolic", "Wake on LAN", chromeWoL},
	{"input-gaming-symbolic", "Total Capture", chromeCapture},
	{"utilities-system-monitor-symbolic", "Connection stats", chromeStats},
	{"view-fullscreen-symbolic", "Toggle fullscreen", chromeFullscreen},
	{"preferences-system-symbolic", "Settings", chromeSettings},
}

type Chrome struct {
	Box    *gtk.Box
	handle *gtk.Image

	app             *Application
	buttons         map[chromeAction]*gtk.Button
	takeBackBtn     *gtk.Button
	vertical        bool

	dragging   bool
	didDrag    bool
	hovering   bool
	marginX    int
	marginY    int
	dragOffX   float64
	dragOffY   float64
	dragStartX int
	dragStartY int
}

// IsDragging reports whether a drag is in progress on the floating menu handle.
// While true, the menu must remain visible (see app.uiAlpha).
func (c *Chrome) IsDragging() bool { return c.dragging }

// IsHovering reports whether the pointer is currently over the floating menu.
// While true, the menu must remain visible.
func (c *Chrome) IsHovering() bool { return c.hovering }

// ClearHover resets the hover state. Called when chrome fades to alpha 0
// since GTK may not deliver a leave event after SetCanTarget(false).
func (c *Chrome) ClearHover() { c.hovering = false }

// HitTest checks if window-relative coordinates (wx, wy) fall within
// the chrome bar's region (with padding for easy discovery).
func (c *Chrome) HitTest(wx, wy float64, windowW, windowH int) bool {
	cw := c.Box.AllocatedWidth()
	ch := c.Box.AllocatedHeight()
	if cw <= 0 || ch <= 0 {
		return false
	}
	const pad = 30
	// Chrome is AlignEnd (right), AlignStart (top), offset by marginEnd/marginTop
	x0 := float64(windowW - c.marginX - cw - pad)
	y0 := float64(c.marginY - pad)
	x1 := float64(windowW - c.marginX + pad)
	y1 := float64(c.marginY + ch + pad)
	return wx >= x0 && wx <= x1 && wy >= y0 && wy <= y1
}

func NewChrome(app *Application) *Chrome {
	c := &Chrome{
		app:     app,
		buttons: make(map[chromeAction]*gtk.Button),
	}

	layout := gtk.OrientationHorizontal
	if app.prefs.ChromeLayout == "vertical" {
		layout = gtk.OrientationVertical
		c.vertical = true
	}
	c.Box = gtk.NewBox(layout, 2)
	c.Box.AddCSSClass("chrome-bar")
	c.Box.SetHAlign(gtk.AlignEnd)
	c.Box.SetVAlign(gtk.AlignStart)

	c.handle = gtk.NewImage()
	c.handle.SetFromIconName("open-menu-symbolic")
	c.handle.SetTooltipText("Drag to move, click to flip H/V")
	c.handle.AddCSSClass("chrome-btn")
	c.handle.AddCSSClass("chrome-handle")
	c.handle.SetCursorFromName("grab")

	c.takeBackBtn = gtk.NewButtonWithLabel("Take Back Control")
	c.takeBackBtn.AddCSSClass("destructive-action")
	c.takeBackBtn.SetVisible(false)
	c.takeBackBtn.ConnectClicked(func() { c.onTakeBackControl() })
	c.Box.Append(c.takeBackBtn)

	for _, def := range chromeButtons {
		btn := gtk.NewButtonFromIconName(def.icon)
		btn.SetTooltipText(def.tooltip)
		btn.AddCSSClass("flat")
		btn.AddCSSClass("chrome-btn")
		c.buttons[def.action] = btn

		action := def.action
		btn.ConnectClicked(func() { c.onAction(action) })
		c.Box.Append(btn)
	}

	// Handle: last in horizontal (right side), first in vertical (top)
	if c.vertical {
		c.Box.Prepend(c.handle)
	} else {
		c.Box.Append(c.handle)
	}

	c.marginX = 8
	c.marginY = 8
	c.Box.SetMarginEnd(c.marginX)
	c.Box.SetMarginTop(c.marginY)

	hover := gtk.NewEventControllerMotion()
	hover.ConnectEnter(func(_, _ float64) {
		c.hovering = true
		c.app.revealUIFor(3 * time.Second)
	})
	hover.ConnectMotion(func(_, _ float64) {
		c.app.revealUIFor(3 * time.Second)
	})
	hover.ConnectLeave(func() {
		c.hovering = false
		c.app.revealUIFor(1500 * time.Millisecond)
	})
	c.Box.AddController(hover)

	return c
}

// AttachToOverlay wires the drag gesture to the parent overlay (a stable,
// non-moving widget). The drag only "claims" if it starts inside the handle's
// bounds. Because the gesture's coordinate system is anchored to the overlay,
// offsetX/Y stay accurate as we move the chrome box during drag (no feedback
// loop, so we can apply position live for visual feedback).
func (c *Chrome) AttachToOverlay(overlay *gtk.Overlay) {
	dragGesture := gtk.NewGestureDrag()
	dragGesture.SetPropagationPhase(gtk.PhaseBubble)

	dragGesture.ConnectDragBegin(func(startX, startY float64) {
		// Convert handle's allocation to overlay coordinates and check
		// whether the press is inside it -- otherwise ignore (let buttons
		// receive the click normally).
		rect, ok := c.handle.ComputeBounds(overlay)
		if !ok {
			return
		}
		if startX < float64(rect.X()) || startX > float64(rect.X()+rect.Width()) ||
			startY < float64(rect.Y()) || startY > float64(rect.Y()+rect.Height()) {
			return
		}
		c.dragging = true
		c.didDrag = false
		c.dragStartX = c.marginX
		c.dragStartY = c.marginY
	})

	dragGesture.ConnectDragUpdate(func(offsetX, offsetY float64) {
		if !c.dragging {
			return
		}
		if !c.didDrag && (absF(offsetX) > 4 || absF(offsetY) > 4) {
			c.didDrag = true
		}
		if !c.didDrag {
			return
		}
		newX := c.dragStartX - int(offsetX)
		newY := c.dragStartY + int(offsetY)
		if newX < 0 {
			newX = 0
		}
		if newY < 0 {
			newY = 0
		}
		c.marginX = newX
		c.marginY = newY
		c.Box.SetMarginEnd(c.marginX)
		c.Box.SetMarginTop(c.marginY)
	})

	dragGesture.ConnectDragEnd(func(offsetX, offsetY float64) {
		if !c.dragging {
			return
		}
		c.dragging = false
		if !c.didDrag {
			c.flipOrientation()
			return
		}
		ww := float64(c.app.window.AllocatedWidth())
		wh := float64(c.app.window.AllocatedHeight())
		if ww > 0 && wh > 0 && c.app.sessionURL != "" {
			pctX := float64(c.marginX) / ww
			pctY := float64(c.marginY) / wh
			c.app.prefs.updateRecent(c.app.sessionURL, func(rc *RecentConnection) {
				rc.ChromePctX = pctX
				rc.ChromePctY = pctY
				rc.ChromeHasPos = true
			})
			savePrefs(c.app.prefs)
		}
	})

	overlay.AddController(dragGesture)
}

func (c *Chrome) flipOrientation() {
	c.vertical = !c.vertical
	if c.vertical {
		c.Box.SetOrientation(gtk.OrientationVertical)
	} else {
		c.Box.SetOrientation(gtk.OrientationHorizontal)
	}
	// Move handle: first in vertical, last in horizontal
	c.Box.Remove(c.handle)
	if c.vertical {
		c.Box.Prepend(c.handle)
	} else {
		c.Box.Append(c.handle)
	}
	layout := "horizontal"
	if c.vertical {
		layout = "vertical"
	}
	c.app.prefs.ChromeLayout = layout
	savePrefs(c.app.prefs)
}

// ApplyPosition restores chrome position from per-device prefs (as % of
// window), falling back to default 8,8 if no saved position exists.
func (c *Chrome) ApplyPosition(sessionURL string) {
	c.marginX = 8
	c.marginY = 8
	if rc := c.app.prefs.findRecent(sessionURL); rc != nil && rc.ChromeHasPos {
		ww := float64(c.app.window.AllocatedWidth())
		wh := float64(c.app.window.AllocatedHeight())
		if ww > 0 && wh > 0 {
			c.marginX = int(rc.ChromePctX * ww)
			c.marginY = int(rc.ChromePctY * wh)
			if c.marginX < 0 {
				c.marginX = 0
			}
			if c.marginY < 0 {
				c.marginY = 0
			}
		}
	}
	c.Box.SetMarginEnd(c.marginX)
	c.Box.SetMarginTop(c.marginY)
}

func (c *Chrome) onAction(action chromeAction) {
	switch action {
	case chromeReconnect:
		if c.app.ctrl != nil {
			c.app.ctrl.ReconnectNow()
		}
	case chromePaste:
		c.app.toggleOverlay("paste")
	case chromeMedia:
		c.app.toggleOverlay("media")
	case chromeSerial:
		c.app.toggleOverlay("serial")
	case chromeWoL:
		c.app.toggleOverlay("wol")
	case chromeCapture:
		c.app.toggleTotalCapture()
	case chromeStats:
		c.app.toggleOverlay("stats")
	case chromeFullscreen:
		c.app.toggleFullscreen()
	case chromeSettings:
		c.app.toggleOverlay("settings")
	}
}

func (c *Chrome) UpdateVisibility(connected, serialActive, captureSupported, otherSession bool) {
	c.buttons[chromePaste].SetVisible(connected)
	c.buttons[chromeMedia].SetVisible(connected)
	c.buttons[chromeSerial].SetVisible(connected && serialActive)
	c.buttons[chromeWoL].SetVisible(connected)
	c.buttons[chromeCapture].SetVisible(connected && captureSupported)
	c.takeBackBtn.SetVisible(otherSession)
}

func (c *Chrome) onTakeBackControl() {
	if c.app.ctrl == nil {
		return
	}
	c.app.ctrl.ReconnectNow()
}

func absF(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
