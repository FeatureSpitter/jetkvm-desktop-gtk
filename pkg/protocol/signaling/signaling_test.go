package signaling

import (
	"encoding/base64"
	"testing"
)

func TestEncodeSDP(t *testing.T) {
	raw := []byte("v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\n")
	encoded := EncodeSDP(raw)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("encoded SDP is not valid base64: %v", err)
	}
	if string(decoded) != string(raw) {
		t.Errorf("round-trip failed: got %q", decoded)
	}
}

func TestDecodeSDP(t *testing.T) {
	original := "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(original))
	decoded, err := DecodeSDP(encoded)
	if err != nil {
		t.Fatalf("DecodeSDP error: %v", err)
	}
	if string(decoded) != original {
		t.Errorf("DecodeSDP = %q, want %q", decoded, original)
	}
}

func TestDecodeSDP_Invalid(t *testing.T) {
	_, err := DecodeSDP("not!valid!base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestSDPRoundTrip(t *testing.T) {
	raw := []byte(`v=0
o=- 46117317 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
m=video 9 UDP/TLS/RTP/SAVPF 96
c=IN IP4 0.0.0.0
a=rtpmap:96 H264/90000`)

	encoded := EncodeSDP(raw)
	decoded, err := DecodeSDP(encoded)
	if err != nil {
		t.Fatalf("round-trip decode: %v", err)
	}
	if string(decoded) != string(raw) {
		t.Error("SDP round-trip mismatch")
	}
}

func TestWebsocketURL(t *testing.T) {
	tests := []struct {
		base string
		want string
		ok   bool
	}{
		{"http://192.168.1.1", "ws://192.168.1.1/webrtc/signaling/client", true},
		{"https://kvm.local", "wss://kvm.local/webrtc/signaling/client", true},
		{"http://10.0.0.1/", "ws://10.0.0.1/webrtc/signaling/client", true},
		{"http://host:8080/prefix", "ws://host:8080/prefix/webrtc/signaling/client", true},
		{"ftp://bad", "", false},
	}
	for _, tt := range tests {
		got, err := WebsocketURL(tt.base)
		if tt.ok {
			if err != nil {
				t.Errorf("WebsocketURL(%q) error: %v", tt.base, err)
				continue
			}
			if got != tt.want {
				t.Errorf("WebsocketURL(%q) = %q, want %q", tt.base, got, tt.want)
			}
		} else {
			if err == nil {
				t.Errorf("WebsocketURL(%q) expected error", tt.base)
			}
		}
	}
}
