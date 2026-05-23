package wol

import (
	"net"
	"testing"
)

func TestParseMAC(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"00:11:22:33:44:55", "00:11:22:33:44:55", true},
		{"00-11-22-33-44-55", "00:11:22:33:44:55", true},
		{"001122334455", "00:11:22:33:44:55", true},
		{"AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff", true},
		{"  00:11:22:33:44:55  ", "00:11:22:33:44:55", true},
		{"", "", false},
		{"not-a-mac", "", false},
		{"00:11:22:33:44", "", false},
		{"ZZZZZZZZZZZZ", "", false},
	}
	for _, tt := range tests {
		hw, err := ParseMAC(tt.input)
		if tt.ok {
			if err != nil {
				t.Errorf("ParseMAC(%q) error: %v", tt.input, err)
				continue
			}
			if hw.String() != tt.want {
				t.Errorf("ParseMAC(%q) = %s, want %s", tt.input, hw, tt.want)
			}
		} else {
			if err == nil {
				t.Errorf("ParseMAC(%q) expected error, got %s", tt.input, hw)
			}
		}
	}
}

func TestFormatMAC(t *testing.T) {
	hw := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	got := FormatMAC(hw)
	want := "00:11:22:33:44:55"
	if got != want {
		t.Errorf("FormatMAC = %q, want %q", got, want)
	}
}

func TestIsValidMAC(t *testing.T) {
	if !IsValidMAC("00:11:22:33:44:55") {
		t.Error("expected valid MAC")
	}
	if IsValidMAC("garbage") {
		t.Error("expected invalid MAC")
	}
	if IsValidMAC("") {
		t.Error("expected empty to be invalid")
	}
}
