package wol

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

const magicPacketSize = 102

// ParseMAC accepts common MAC address formats (colon, dash, or bare hex)
// and returns a 6-byte hardware address.
func ParseMAC(s string) (net.HardwareAddr, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty MAC address")
	}
	hw, err := net.ParseMAC(s)
	if err == nil && len(hw) == 6 {
		return hw, nil
	}
	// Try bare hex (e.g. "001122334455").
	bare := strings.ReplaceAll(strings.ReplaceAll(s, ":", ""), "-", "")
	if len(bare) == 12 {
		b, err := hex.DecodeString(bare)
		if err == nil && len(b) == 6 {
			return net.HardwareAddr(b), nil
		}
	}
	return nil, fmt.Errorf("invalid MAC address: %s", s)
}

// FormatMAC returns the canonical colon-separated lowercase representation.
func FormatMAC(hw net.HardwareAddr) string {
	return hw.String()
}

// buildPacket constructs the 102-byte WOL magic packet.
func buildPacket(mac net.HardwareAddr) [magicPacketSize]byte {
	var packet [magicPacketSize]byte
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], mac)
	}
	return packet
}

// broadcastAddrs returns the directed broadcast address for every IPv4
// interface, plus the global 255.255.255.255 as a fallback.
func broadcastAddrs() []net.IP {
	seen := map[string]bool{}
	var addrs []net.IP

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifAddrs, _ := iface.Addrs()
		for _, a := range ifAddrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			ip4 := ipNet.IP.To4()
			mask := ipNet.Mask
			bcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				bcast[i] = ip4[i] | ^mask[i]
			}
			key := bcast.String()
			if !seen[key] {
				seen[key] = true
				addrs = append(addrs, bcast)
			}
		}
	}

	if !seen[net.IPv4bcast.String()] {
		addrs = append(addrs, net.IPv4bcast)
	}
	return addrs
}

// Send broadcasts a Wake-on-LAN magic packet for the given MAC address
// on every network interface's broadcast address (port 9).
func Send(mac net.HardwareAddr) error {
	if len(mac) != 6 {
		return fmt.Errorf("MAC address must be 6 bytes, got %d", len(mac))
	}
	packet := buildPacket(mac)
	targets := broadcastAddrs()
	var firstErr error
	sent := 0
	for _, ip := range targets {
		conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: ip, Port: 9})
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("dial %s: %w", ip, err)
			}
			continue
		}
		_, err = conn.Write(packet[:])
		conn.Close()
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("send to %s: %w", ip, err)
			}
			continue
		}
		sent++
	}
	if sent == 0 && firstErr != nil {
		return firstErr
	}
	return nil
}

// IsValidMAC returns true if s can be parsed as a 6-byte MAC address.
func IsValidMAC(s string) bool {
	_, err := ParseMAC(s)
	return err == nil
}
