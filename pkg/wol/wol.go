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

type sendTarget struct {
	localIP   net.IP // bind to this local address (nil = OS default)
	destIP    net.IP // send to this broadcast address
	ifaceName string // for logging
}

// sendTargets enumerates all IPv4 interfaces and returns targets that
// cover every possible path: directed broadcast per subnet AND global
// broadcast bound to each interface's local IP (covers /32 point-to-point
// VPNs like WireGuard/Tailscale where directed broadcast equals the IP).
func sendTargets() []sendTarget {
	type key struct{ local, dest string }
	seen := map[key]bool{}
	var targets []sendTarget
	add := func(t sendTarget) {
		k := key{t.localIP.String(), t.destIP.String()}
		if !seen[k] {
			seen[k] = true
			targets = append(targets, t)
		}
	}

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

			// Directed broadcast for this subnet.
			bcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				bcast[i] = ip4[i] | ^mask[i]
			}
			add(sendTarget{localIP: ip4, destIP: bcast, ifaceName: iface.Name})

			// Global broadcast bound to this interface's IP — ensures
			// the packet goes out this specific interface even for /32.
			add(sendTarget{localIP: ip4, destIP: net.IPv4bcast, ifaceName: iface.Name})
		}
	}

	// Fallback: unbound global broadcast (OS picks interface).
	if len(targets) == 0 {
		targets = append(targets, sendTarget{destIP: net.IPv4bcast, ifaceName: "default"})
	}
	return targets
}

// Send broadcasts a Wake-on-LAN magic packet for the given MAC address
// on every network interface (port 9), using both directed and global
// broadcast to cover LAN, VPN, and point-to-point interfaces.
func Send(mac net.HardwareAddr) error {
	if len(mac) != 6 {
		return fmt.Errorf("MAC address must be 6 bytes, got %d", len(mac))
	}
	packet := buildPacket(mac)
	targets := sendTargets()
	var firstErr error
	sent := 0
	for _, t := range targets {
		var local *net.UDPAddr
		if t.localIP != nil {
			local = &net.UDPAddr{IP: t.localIP}
		}
		conn, err := net.DialUDP("udp4", local, &net.UDPAddr{IP: t.destIP, Port: 9})
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("dial %s→%s (%s): %w", t.localIP, t.destIP, t.ifaceName, err)
			}
			continue
		}
		_, err = conn.Write(packet[:])
		conn.Close()
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("send %s→%s (%s): %w", t.localIP, t.destIP, t.ifaceName, err)
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
