# Manual Release Checklist

Automated tests cover the emulator-backed paths. Before each release,
verify these items on a **real JetKVM device** connected to a live system.

## Connection

- [ ] mDNS/HTTP probe discovers device on local network
- [ ] Direct-connect via CLI `jetkvm-desktop 192.168.x.x` works
- [ ] Password authentication prompt appears when required
- [ ] Wrong password shows "Incorrect password", allows retry
- [ ] Reconnect after pulling/replugging Ethernet cable succeeds

## Video

- [ ] Live video displays with correct aspect ratio
- [ ] VA-API hardware decoding works (check `Connection stats` overlay)
- [ ] Fullscreen toggle (F11 or chrome button) works 3x without glitch
- [ ] Stream quality High/Medium/Low visibly changes compression

## Keyboard

- [ ] All alphanumeric keys on en-US layout map correctly
- [ ] All alphanumeric keys on pt-PT layout map correctly
- [ ] Modifier keys (Ctrl, Alt, Shift, Super) work in combinations
- [ ] Paste text with accented characters (e.g. "Olá Ação") on pt-PT
- [ ] Total Capture mode intercepts Alt+Tab, sends it to remote
- [ ] Total Capture toggle key (ScrollLock by default) works to enter/exit

## Mouse

- [ ] Absolute pointer tracks correctly at all corners
- [ ] Left/middle/right buttons register on remote
- [ ] Scroll wheel works in both directions
- [ ] Invert scroll pref actually inverts
- [ ] Hide cursor pref blanks the local cursor over the video area
- [ ] Side buttons (back/forward) register when pref is enabled

## Virtual Media

- [ ] Mount ISO from URL — remote boots from it
- [ ] Upload local ISO — appears in storage list
- [ ] Mount uploaded file — remote sees it as USB drive
- [ ] Unmount — remote no longer sees the drive

## Serial Console

- [ ] Serial console output scrolls in real time
- [ ] Typing in serial input sends characters to device
- [ ] Serial settings (baud rate, parity) can be changed

## Wake on LAN

- [ ] Add WoL device with MAC and broadcast IP
- [ ] Send WoL packet — target machine wakes

## Settings

- [ ] All settings panels open and close without crash
- [ ] Chrome bar drag-to-reposition persists across reconnect
- [ ] Chrome bar H/V flip persists
- [ ] Theme switch (system/light/dark) applies immediately
- [ ] EDID presets apply without error

## Other Session

- [ ] When second client connects, first sees "Take Back Control" button
- [ ] Clicking "Take Back Control" reconnects and kicks the other client

## Edge Cases

- [ ] Device reboot (via settings) — app shows rebooting state, reconnects
- [ ] Close and reopen app — recent connections list preserved
- [ ] App exit during Total Capture — compositor shortcuts restored
