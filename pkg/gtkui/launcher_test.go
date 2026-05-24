package gtkui

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"192.168.1.50", "https://192.168.1.50"},
		{"jetkvm.local", "https://jetkvm.local"},
		{"http://jetkvm.local", "http://jetkvm.local"},
		{"https://10.0.0.1", "https://10.0.0.1"},
	}
	for _, tt := range tests {
		got := normalizeURL(tt.in)
		if got != tt.want {
			t.Errorf("normalizeURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsValidHost(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"192.168.1.50", true},
		{"jetkvm.local", true},
		{"", false},
		{"  ", false},
		{"http://ok.com", true},
	}
	for _, tt := range tests {
		got := isValidHost(tt.in)
		if got != tt.want {
			t.Errorf("isValidHost(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "just now"},
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{2 * time.Hour, "2h ago"},
		{48 * time.Hour, "2d ago"},
	}
	for _, tt := range tests {
		got := formatAge(tt.d)
		if got != tt.want {
			t.Errorf("formatAge(%s) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestPrefsRecentRoundtrip(t *testing.T) {
	p := defaultPrefs()
	p.addRecent("https://a.local", "A")
	// Windows has ~15ms timer resolution; ensure distinct timestamps.
	time.Sleep(20 * time.Millisecond)
	p.addRecent("https://b.local", "B")
	time.Sleep(20 * time.Millisecond)
	p.addRecent("https://c.local", "C")

	if len(p.RecentConnections) != 3 {
		t.Fatalf("got %d recents, want 3", len(p.RecentConnections))
	}
	if p.RecentConnections[0].URL != "https://c.local" {
		t.Errorf("first recent = %q, want c.local", p.RecentConnections[0].URL)
	}

	p.removeRecent("https://b.local")
	if len(p.RecentConnections) != 2 {
		t.Fatalf("after remove, got %d recents, want 2", len(p.RecentConnections))
	}
}

func TestPrefsLegacyHideHeaderBarIgnored(t *testing.T) {
	raw := `{"hide_header_bar": true, "theme": "dark", "pin_chrome": true}`
	var p Preferences
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal legacy prefs: %v", err)
	}
	if p.Theme != "dark" {
		t.Errorf("theme = %q, want dark", p.Theme)
	}
	if !p.PinChrome {
		t.Error("pin_chrome should be true")
	}
}

func TestPrefsDefaults(t *testing.T) {
	p := defaultPrefs()
	if p.PointerMoveThrottleMs != 8 {
		t.Errorf("PointerMoveThrottleMs = %d, want 8", p.PointerMoveThrottleMs)
	}
	if !p.AbsoluteSideButtonsViaRel {
		t.Error("AbsoluteSideButtonsViaRel should be true by default")
	}
	if p.ConnectWindowMode != "maximize" {
		t.Errorf("ConnectWindowMode = %q, want maximize", p.ConnectWindowMode)
	}
}

func TestMouseButton(t *testing.T) {
	tests := []struct {
		gtk  uint
		want byte
	}{
		{1, 1},  // left
		{2, 4},  // middle
		{3, 2},  // right
		{8, 8},  // side back
		{9, 16}, // side forward
		{0, 0},  // unknown
		{99, 0}, // unknown
	}
	for _, tt := range tests {
		got := mouseButton(tt.gtk)
		if got != tt.want {
			t.Errorf("mouseButton(%d) = %d, want %d", tt.gtk, got, tt.want)
		}
	}
}
