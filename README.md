# jetkvm-desktop-gtk

A native GTK4 desktop client for [JetKVM](https://jetkvm.com) with GPU-accelerated video rendering, designed to run smoothly on older and low-power hardware.

> **Fork of [lkarlslund/jetkvm-desktop](https://github.com/lkarlslund/jetkvm-desktop)**
> — the original cross-platform JetKVM desktop client built with [Ebiten](https://ebitengine.org/).
> This fork replaces the Ebiten rendering layer with native GTK4 + OpenGL ES 3.0
> to offload video compositing to the GPU. All credit for the original architecture,
> protocol implementation, and feature set goes to
> [@lkarlslund](https://github.com/lkarlslund) and the upstream contributors.

## Why this fork?

The original Ebiten-based client decodes and composites video frames on the CPU, which works well on modern hardware but becomes **unusable on older machines** — CPU pegs at 100%, the UI lags, and battery drains rapidly. Meanwhile, the browser-based JetKVM client handles the same stream effortlessly, proving the bottleneck is in the desktop app's rendering pipeline.

This fork gives old laptops and low-power machines a second life as **responsive KVM terminals** — a 10-year-old laptop with integrated graphics can act as a zero-lag viewer for any machine behind a JetKVM.

### Key differences from upstream

| | Upstream (Ebiten) | This fork (GTK4) |
|---|---|---|
| **Rendering** | CPU-based frame compositing | GPU-accelerated OpenGL ES 3.0 shaders |
| **UI toolkit** | Custom-drawn Ebiten widgets | Native GTK4 components |
| **CPU usage** | High on older hardware | Minimal — offloaded to GPU |
| **Platforms** | Windows, macOS, Linux | Linux (X11) — other platforms planned |

## Features

Full feature parity with the upstream client, plus additional improvements:

- **GPU-accelerated video** — YCbCr → RGB conversion via OpenGL ES 3.0 fragment shaders
- **Native GTK4 UI** — launcher, settings, overlays, floating menu, theme support (dark/light/system)
- **Total input capture** — X11 keyboard grab with direct HID scancode forwarding (ScrollLock toggle, configurable)
- **Prioritized network discovery** — physical NICs first, then VPNs/containers/Docker networks
- **Configurable paste** — per-keystroke delay, 19 keyboard layouts, unsupported character preview
- **Wake-on-LAN** — overlay with device management via KVM RPC
- **Remote hotkeys** — experimental Alt+Tab forwarding via chord shortcuts
- **Mouse side buttons** — back/forward (buttons 8/9) supported natively

![jetkvm-desktop launcher](docs/launcher.png)

## Getting Started

### Build

```bash
# Dependencies (Debian/Ubuntu)
sudo apt install libgtk-4-dev libglib2.0-dev build-essential

# Build
CGO_ENABLED=1 go build -o jetkvm-desktop-gtk ./cmd/jetkvm-desktop
```

### Run

```bash
# Open launcher with device discovery
./jetkvm-desktop-gtk

# Connect directly to a device
./jetkvm-desktop-gtk jetkvm.local
./jetkvm-desktop-gtk 192.168.1.50
```

If the device requires a password, the app will prompt for it.

![jetkvm-desktop settings](docs/settings.png)

## Platform Support

**Linux** (GTK4 + X11) is the primary development target with full feature support including X11 window centering and total input capture via keyboard grabs.

**macOS** and **Windows** are supported via GTK4's native backends. Total input capture uses platform-specific APIs (CGEventTap on macOS, low-level hooks on Windows). Window centering falls back to GTK4's default placement on these platforms.

Wayland support depends on GTK4's Wayland backend maturity and input capture protocol availability.

## Attribution

This project is a fork of [lkarlslund/jetkvm-desktop](https://github.com/lkarlslund/jetkvm-desktop) by [@lkarlslund](https://github.com/lkarlslund). The original project provides the protocol layer, session management, device discovery, input handling, and the full feature set that this fork builds upon. The upstream maintainer's decision to keep Ebiten as the cross-platform rendering foundation is entirely reasonable — this fork simply explores a different trade-off optimized for GPU-accelerated rendering on Linux.

For the upstream JetKVM ecosystem, see [github.com/jetkvm](https://github.com/jetkvm).

## License

Same license as the upstream project.
