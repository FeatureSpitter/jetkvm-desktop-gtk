package gtkui

import (
	"context"
	"log"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/FeatureSpitter/jetkvm-desktop-gtk/pkg/session"
	"github.com/FeatureSpitter/jetkvm-desktop-gtk/pkg/wol"
)

// WoLOverlay manages Wake-on-LAN saved devices and sending magic packets.
type WoLOverlay struct {
	Box *gtk.Box

	app        *Application
	deviceList *gtk.ListBox
	nameEntry  *gtk.Entry
	macEntry   *gtk.Entry
	addBtn     *gtk.Button
	statusLabel *gtk.Label
	closeBtn   *gtk.Button

	devices []session.WakeOnLanDevice
}

func NewWoLOverlay(app *Application) *WoLOverlay {
	w := &WoLOverlay{app: app}
	w.Box = gtk.NewBox(gtk.OrientationVertical, 8)
	w.Box.AddCSSClass("overlay-panel")
	w.Box.SetHAlign(gtk.AlignCenter)
	w.Box.SetVAlign(gtk.AlignCenter)

	title := gtk.NewLabel("Wake on LAN")
	title.AddCSSClass("title-3")
	title.SetXAlign(0)
	w.Box.Append(title)

	desc := gtk.NewLabel("Send Wake-on-LAN magic packets to start remote machines.")
	desc.SetWrap(true)
	desc.AddCSSClass("dim-label")
	desc.SetXAlign(0)
	w.Box.Append(desc)

	// Saved devices
	w.deviceList = gtk.NewListBox()
	w.deviceList.SetSelectionMode(gtk.SelectionNone)
	w.deviceList.AddCSSClass("boxed-list")

	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(w.deviceList)
	scroll.SetMinContentHeight(100)
	scroll.SetMaxContentHeight(200)
	w.Box.Append(scroll)

	// Add new device
	addLabel := gtk.NewLabel("Add New Device")
	addLabel.AddCSSClass("title-4")
	addLabel.SetXAlign(0)
	w.Box.Append(addLabel)

	w.nameEntry = gtk.NewEntry()
	w.nameEntry.SetPlaceholderText("Device name")
	w.Box.Append(w.nameEntry)

	w.macEntry = gtk.NewEntry()
	w.macEntry.SetPlaceholderText("MAC address (e.g. AA:BB:CC:DD:EE:FF)")
	w.Box.Append(w.macEntry)

	w.addBtn = gtk.NewButtonWithLabel("+ Add Device")
	w.addBtn.AddCSSClass("suggested-action")
	w.addBtn.ConnectClicked(func() { w.addDevice() })
	w.Box.Append(w.addBtn)

	w.statusLabel = gtk.NewLabel("")
	w.statusLabel.AddCSSClass("dim-label")
	w.statusLabel.SetXAlign(0)
	w.Box.Append(w.statusLabel)

	w.closeBtn = gtk.NewButtonWithLabel("Close")
	w.closeBtn.ConnectClicked(func() { app.closeOverlay() })
	w.Box.Append(w.closeBtn)

	return w
}

func (w *WoLOverlay) Refresh() {
	tw, th := overlayTargetSize(w.app, 560, 520)
	w.Box.SetSizeRequest(tw, th)
	if w.app.ctrl == nil {
		log.Printf("[wol] refresh: no controller, skipping")
		return
	}
	devices, err := w.app.ctrl.GetWakeOnLanDevices(context.Background())
	if err != nil {
		log.Printf("[wol] error loading devices: %v", err)
		w.statusLabel.SetText("Error loading devices: " + err.Error())
		return
	}
	log.Printf("[wol] loaded %d device(s)", len(devices))
	w.devices = devices
	w.rebuildList()
}

func (w *WoLOverlay) rebuildList() {
	removeAllChildren(w.deviceList)

	for _, dev := range w.devices {
		row := w.makeDeviceRow(dev)
		w.deviceList.Append(row)
	}
}

func (w *WoLOverlay) makeDeviceRow(dev session.WakeOnLanDevice) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.SetMarginStart(8)
	box.SetMarginEnd(8)

	nameLabel := gtk.NewLabel(dev.Name)
	nameLabel.SetHExpand(true)
	nameLabel.SetXAlign(0)

	macLabel := gtk.NewLabel(dev.MacAddress)
	macLabel.AddCSSClass("dim-label")

	wakeBtn := gtk.NewButtonWithLabel("Wake")
	wakeBtn.AddCSSClass("suggested-action")
	mac := dev.MacAddress
	wakeBtn.ConnectClicked(func() { w.wakeDevice(mac) })

	delBtn := gtk.NewButtonFromIconName("window-close-symbolic")
	delBtn.AddCSSClass("flat")
	devName := dev.Name
	delBtn.ConnectClicked(func() { w.deleteDevice(devName) })

	box.Append(nameLabel)
	box.Append(macLabel)
	box.Append(wakeBtn)
	box.Append(delBtn)

	row := gtk.NewListBoxRow()
	row.SetChild(box)
	return row
}

func (w *WoLOverlay) wakeDevice(mac string) {
	log.Printf("[wol] sending magic packet to %s", mac)

	hw, err := wol.ParseMAC(mac)
	if err != nil {
		log.Printf("[wol] invalid MAC %s: %v", mac, err)
		w.statusLabel.SetText("Invalid MAC: " + err.Error())
		return
	}
	if err := wol.Send(hw); err != nil {
		log.Printf("[wol] local send error for %s: %v", mac, err)
		w.statusLabel.SetText("WOL " + mac + " — local: " + err.Error())
		return
	}
	log.Printf("[wol] local magic packet sent to %s (all interfaces)", mac)

	ctrl := w.app.ctrl
	if ctrl == nil {
		w.statusLabel.SetText("Magic packet sent to " + mac)
		return
	}

	w.statusLabel.SetText("Magic packet sent to " + mac)
	go func() {
		err := ctrl.SendWakeOnLan(mac, "")
		glib.IdleAdd(func() {
			if err != nil {
				log.Printf("[wol] remote send error for %s (non-fatal): %v", mac, err)
			} else {
				log.Printf("[wol] remote magic packet also sent via KVM for %s", mac)
			}
		})
	}()

	glib.TimeoutAdd(5000, func() bool {
		if w.app.ctrl != nil {
			log.Printf("[wol] auto-reconnecting to pick up new video after wake")
			w.app.ctrl.ReconnectNow()
		}
		return false
	})
}

func (w *WoLOverlay) addDevice() {
	name := w.nameEntry.Text()
	mac := w.macEntry.Text()
	if name == "" || mac == "" {
		return
	}
	w.devices = append(w.devices, session.WakeOnLanDevice{
		Name:       name,
		MacAddress: mac,
	})
	w.saveDevices()
	w.nameEntry.SetText("")
	w.macEntry.SetText("")
	w.rebuildList()
}

func (w *WoLOverlay) deleteDevice(name string) {
	var filtered []session.WakeOnLanDevice
	for _, d := range w.devices {
		if d.Name != name {
			filtered = append(filtered, d)
		}
	}
	w.devices = filtered
	w.saveDevices()
	w.rebuildList()
}

func (w *WoLOverlay) saveDevices() {
	if w.app.ctrl == nil {
		return
	}
	items := make([]struct {
		Name        string
		MacAddress  string
		BroadcastIP string
	}, len(w.devices))
	for i, d := range w.devices {
		items[i].Name = d.Name
		items[i].MacAddress = d.MacAddress
		items[i].BroadcastIP = d.BroadcastIP
	}
	_ = w.app.ctrl.SetWakeOnLanDevices(items)
}
