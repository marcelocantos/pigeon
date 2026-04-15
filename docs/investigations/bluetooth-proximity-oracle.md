# Bluetooth as a Proximity Oracle

**Status:** Investigation  
**Target:** 🎯T5  
**Date:** 2026-04-15

---

## Motivation

Pigeon's pairing ceremony uses a 6-digit confirmation code to detect man-in-the-middle attacks — the user visually compares the code shown on both devices and approves it. This is a human-in-the-loop step that works but has friction: the user must read a code and type it (or tap confirm).

Bluetooth Low Energy (BLE) can sense physical proximity. Two devices that are physically near each other can exchange or detect BLE signals. This investigation evaluates whether BLE proximity can:

1. **Supplement the confirmation code** — proximity evidence as a second factor ("the devices are near each other AND the code matches").
2. **Replace the confirmation code** for low-threat environments — proximity alone as pairing confirmation.
3. **Gate transport upgrades** — only use LAN-direct or hole-punched P2P when devices are physically near each other (preventing remote relaying from impersonating a local session).

This document focuses exclusively on proximity sensing. BLE is not evaluated as a data channel.

---

## BLE Fundamentals

Bluetooth Low Energy (BLE, Bluetooth 4.0+) has two relevant primitives:

- **Advertising:** A device broadcasts small packets (31–255 bytes) at configurable intervals. No pairing required. Any scanner in range receives them.
- **Scanning:** A device listens for advertising packets and receives RSSI (signal strength) measurements for each observed advertiser.

RSSI is a proxy for distance: stronger signal = closer. RSSI is noisy (±10 dBm is common), but at very short range (< 1 metre) signals are reliably strong (> −60 dBm), and at long range (> 10 metres) they are reliably weak (< −80 dBm). This makes RSSI suitable for a coarse proximity gate (near / not near) rather than precise distance measurement.

**References:**
- Bluetooth Core Specification 5.4: https://www.bluetooth.com/specifications/specs/core-specification-5-4/
- Apple CoreBluetooth framework: https://developer.apple.com/documentation/corebluetooth
- Android BluetoothLeAdvertiser: https://developer.android.com/reference/android/bluetooth/le/BluetoothLeAdvertiser

---

## iOS Platform: CoreBluetooth

### Advertising

`CBPeripheralManager` can advertise a custom service UUID and arbitrary data. Limitations:

- **Background advertising is heavily throttled.** When the app is backgrounded, iOS restricts BLE advertising to a system-controlled interval (typically 1–4 seconds, up from the ~20ms foreground default). The service UUID is still broadcast but data payload is stripped to a service UUID list only.
- **No raw advertisement data control.** iOS does not allow setting arbitrary manufacturer-specific data in the advertisement payload from third-party apps. Only the `CBAdvertisementDataLocalNameKey` and service UUIDs are settable.
- **Local name is 10 bytes max** (system truncates).

### Scanning

`CBCentralManager` can scan for advertisements by service UUID. RSSI is reported for each advertisement via `-centralManager:didDiscoverPeripheral:advertisementData:RSSI:`.

- **Background scanning:** iOS permits background BLE scanning only for apps with the `bluetooth-central` background mode declared in `Info.plist` and entitlement. Even then, iOS coalesces scan results in the background to reduce power usage.
- **Foreground scanning:** Full RSSI and advertisement data available with no restrictions beyond user-granted Bluetooth permission (required since iOS 13, `NSBluetoothAlwaysUsageDescription`).

### Proximity-specific API: Core Location (iBeacon)

Apple's **iBeacon** protocol is a BLE-based proximity API built on CoreBluetooth. It broadcasts a UUID, major, and minor value; `CLLocationManager` reports proximity as `.immediate`, `.near`, `.far`, or `.unknown`.

- **iBeacon ranging** works in the background (with `location` background mode, not `bluetooth-central`).
- The proximity categories map to approximate distances: immediate < 0.5m, near < 3m, far > 3m.
- This is purpose-built for the proximity use case.

**Reference:** https://developer.apple.com/documentation/corelocation/ranging_for_beacons

### Nearby Interaction framework (UWB)

iOS 14+ on devices with the U1 chip (iPhone 11+) supports **Ultra-Wideband** ranging via `NINearbyObject`. This provides centimetre-precision distance + direction. However:
- Requires both devices to run a Nearby Interaction session (app must be active on both).
- Not available on iPad (as of 2026) or older iPhones.
- BLE is used as the discovery channel for Nearby Interaction sessions.

**Reference:** https://developer.apple.com/documentation/nearbyinteraction

---

## Android Platform: BluetoothLeAdvertiser

### Advertising

`BluetoothLeAdvertiser` (Android 5.0+) supports:

- **Manufacturer-specific data** in the advertisement payload — unlike iOS, Android apps can embed arbitrary bytes (up to 26 bytes in the manufacturer data field of a legacy advertisement; up to ~225 bytes with extended advertising on Android 8+).
- **Background advertising** via a `ForegroundService` with a notification. Without a foreground service, Android may kill the advertising process.
- **Advertising sets** (Android 8+): multiple simultaneous advertisement sets with different parameters.

### Scanning

`BluetoothLeScanner` with `ScanFilter` by service UUID or manufacturer ID. RSSI is reported per scan result.

- **Background scanning:** requires `ACCESS_BACKGROUND_LOCATION` (Android 10+) — a notably invasive permission for a BLE proximity check. This is a significant UX and privacy friction point.
- **Foreground scanning:** only `BLUETOOTH_SCAN` permission needed (Android 12+).

### Nearby Connections API (Google Play Services)

Google's **Nearby Connections** API provides a higher-level proximity and transport primitive. It automatically selects BLE, Wi-Fi Direct, Hotspot, or Bluetooth Classic based on what is available and uses RSSI and connection quality to infer proximity.

However:
- Requires Google Play Services — not available on AOSP/custom Android.
- The API is designed for data transfer, not pure proximity sensing. Pigeon would need only the discovery layer.

**Reference:** https://developers.google.com/nearby/connections/overview

---

## Power and Battery Implications

### BLE advertising

Advertising at the minimum interval (100ms) consumes approximately 0.5–1.5 mA on a modern SoC (Bluetooth 5 hardware). At a battery-friendly 1-second interval: ~0.15–0.5 mA average. This is negligible for brief pairing events but non-trivial for continuous background advertising.

### BLE scanning

Scanning is more expensive than advertising. Continuous scan: 10–20 mA. Duty-cycled scan (scan 100ms / sleep 1s): ~1–2 mA. iOS and Android both impose duty cycling for background scans to protect battery.

### Assessment for pigeon

Proximity sensing is needed only during the pairing ceremony, which is a short (< 60 second) interactive event. The user has the app in the foreground. Foreground BLE scan during pairing adds < 1% battery overhead even on a short 10% battery charge level. This is entirely acceptable.

The concern is if proximity is used as an ongoing session guard (e.g., "downgrade to relay when devices separate"). Continuous background proximity monitoring would be expensive and is not recommended.

**Conclusion:** Use proximity only during the pairing ceremony (foreground, time-limited). Do not use continuous background proximity monitoring.

---

## MAC Address Rotation and Peer Identity

### The rotation problem

Both iOS and Android rotate BLE MAC addresses periodically to prevent long-term tracking:

- **iOS:** Rotates approximately every 15 minutes, or when the app moves to background, or on network change. The rotated address is random.
- **Android:** Rotates on each advertising interval start (since Android 10 with `AdvertisingSet`), or periodically (behaviour varies by vendor).

**Consequence:** MAC address cannot be used as a stable peer identifier. A device that was seen at address `AA:BB:CC:DD:EE:FF` ten minutes ago may now advertise from `11:22:33:44:55:66`.

### Solutions

**Service UUID as identity anchor**  
Include a session-specific UUID in the BLE advertisement (iOS: service UUID; Android: service UUID or manufacturer data). This UUID is derived from the ongoing pairing ceremony and identifies the peer without depending on MAC address. Both sides generate and exchange this UUID early in the ceremony (e.g., after ECDH but before confirmation code display). The UUID is short-lived (only valid for this pairing session) so it does not create a long-term tracking vector.

**Hash of ECDH public key**  
After the ECDH exchange, both peers know both public keys. A truncated hash of the peer's public key can serve as the BLE service UUID (or manufacturer data prefix). This is self-authenticating — the scanner confirms proximity to the peer it is already talking to, not an arbitrary device. MAC rotation does not matter because the service UUID is the identity.

**Reference:** Apple's recommendation for privacy-preserving BLE: https://developer.apple.com/videos/play/wwdc2021/10090/

---

## Integration with the Pairing Ceremony

The current ceremony:

```
cli → server: pair_begin
server → ios:  [relay: ECDH key exchange]
               [both sides compute shared secret + 6-digit code]
user:          visually compares codes on both screens → approves
server → ios:  [relay: auth token exchange]
```

### Option A: Proximity as confirmation-code supplement

After ECDH (both sides know both pubkeys):

1. Each device derives a short-lived BLE service UUID from the ECDH session (e.g., HKDF output truncated to 16 bytes, formatted as UUID).
2. Each device starts BLE advertising with this UUID and simultaneously scans for it.
3. If both devices detect each other above an RSSI threshold (e.g., > −70 dBm = within ~3 metres), a proximity confirmation is generated.
4. The UI shows: "Devices are physically near each other" ✓ (or "Could not verify proximity").
5. The user still confirms the code (proximity is additive evidence, not a replacement).

**Security value:** An attacker performing a relay MITM must be physically present in the room. This significantly raises the attack cost.

**Ceremony change:** Adds a parallel BLE scan step (~3–5 seconds) after ECDH. The ceremony is extended slightly but the user experience improves (proximity confirmation is automatic, no code reading required if both checks pass).

### Option B: Proximity as confirmation-code replacement (low-threat mode)

Same as Option A, but if proximity is confirmed above a threshold (e.g., > −60 dBm = within ~1 metre), skip the code display entirely. Auto-confirm.

**Security value:** Weaker than Option A. An attacker who is physically present (within 1 metre) can still perform a relay MITM (though this is a very high-cost attack). Suitable for low-sensitivity applications where pairing friction is more important to minimise than MitM risk.

**Ceremony change:** Code display step becomes conditional. Adds complexity to the state machine.

### Option C: Proximity as LAN-upgrade gate

After pairing is complete, when pigeon considers upgrading from relay to LAN-direct: require BLE proximity as a condition. This prevents a remote attacker who has compromised the relay from impersonating a local session upgrade.

**Security value:** Moderate. Adds defence-in-depth for transport upgrades.

**Ceremony change:** None to the pairing ceremony. Adds a new "proximity check" step to the session lifecycle.

---

## Privacy Considerations

### BLE advertisements are visible to all nearby scanners

During the proximity check, both devices advertise a BLE service UUID derived from the ECDH session. This UUID is:
- **Session-specific and short-lived** (valid only during pairing).
- **Not linkable** to the device identity or prior sessions.
- **Visible to any BLE scanner in range** — a nearby device could log "this UUID was advertised here at this time."

The UUID itself reveals nothing about the devices or users. The timing and location of the advertisement could be logged by a passive scanner, but this applies to all BLE activity and is not specific to pigeon.

**Mitigations:**
- Minimise advertisement duration (start BLE only when the pairing UI is open, stop as soon as proximity is confirmed or the ceremony ends).
- Use a UUID with no identifiable structure (random HKDF output, not an obvious pigeon-branded UUID).

### Android background location permission

Android's `ACCESS_BACKGROUND_LOCATION` requirement for background BLE scanning is a significant privacy-optics issue. Since pigeon only needs foreground proximity sensing during an active pairing flow, this permission should not be requested. Pairing should require the app to be in the foreground — this is already the expected UX.

### MAC address rotation (revisited)

From the privacy side, MAC rotation is a feature, not a problem. Pigeon's BLE identity is the session-derived UUID, not the MAC address. No persistent BLE identity is created.

---

## Recommendation

**Pursue Option A (proximity as supplement to confirmation code) as part of the pairing ceremony polish.**

The engineering cost is moderate (~2–3 weeks for Go signalling + iOS CoreBluetooth + Android BluetoothLeAdvertiser). The security improvement is concrete: MITM attacks must become physical attacks. The UX improvement is also concrete: if proximity is confirmed automatically, the confirmation code display can be de-emphasised ("Your devices confirmed they are near each other — confirm the codes match as a final check") or the confirm tap simplified.

**Do not pursue Option B (auto-confirm)** unless a specific low-sensitivity use case emerges that justifies the weaker security model.

**Option C (proximity-gated LAN upgrade)** is worth revisiting when LAN-direct transport is implemented, but it is out of scope for the current pairing ceremony work.

**Key constraint:** Foreground-only. Do not request background Bluetooth permissions. The pairing ceremony is already a foreground-interactive flow.

---

## Open Questions

1. **iBeacon vs raw CoreBluetooth:** Should pigeon use CoreLocation/iBeacon (simpler, backgroundable) or raw CoreBluetooth advertising (more control)? iBeacon ranging is purpose-built for this but requires a `location` background mode which may be surprising to users.
2. **Android foreground service requirement:** Is a foreground service (with its mandatory notification) acceptable UX for BLE advertising during pairing? Or is scan-only (no advertising) from the Android side sufficient?
3. **RSSI threshold calibration:** What RSSI threshold correctly distinguishes "same room" from "adjacent room"? This requires empirical testing across device pairs (iPhone ↔ Android, iPhone ↔ Mac, etc.).
4. **BLE unavailable:** Some devices may have Bluetooth disabled or unavailable (airplane mode, low-power mode). Proximity sensing must degrade gracefully to "proximity unknown" without blocking the ceremony.
5. **Nearby Interaction (UWB):** For iPhone 11+ pairs, UWB provides far more accurate proximity than BLE. Should pigeon use UWB when available and fall back to BLE? This adds platform-detection complexity.
6. **State machine changes:** The YAML-defined state machine would need new states for "awaiting proximity" and "proximity confirmed". How much does this complicate the machine, and does it stay within the existing framework's expressiveness?
7. **Android 12+ permission model:** `BLUETOOTH_SCAN` without `ACCESS_FINE_LOCATION` (Android 12+) requires `neverForLocation` flag. Does this restrict what scan data is available? Does it affect service UUID filtering?
