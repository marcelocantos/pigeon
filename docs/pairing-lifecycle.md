# Pigeon — Pairing Artifact Lifecycle

This guide walks through the full credential lifecycle that pigeon
exposes for application-level pairing: a server mints a
`PairingArtifact`, the artifact reaches the client device through one
of two delivery channels (QR scan or developer-deploy via `xcrun`),
the client persists it via a `CredentialStore`, reconnects through it
on every launch, and re-pairs when the artifact expires.

The primitives are wire-compatible across Go, Swift, and Kotlin: an
artifact minted in any SDK can be decoded by the other two. This guide
shows side-by-side code in all three.

## Concepts at a glance

- **`PairingRecord`** — the cryptographic core of a paired peer:
  the peer's instance ID, the relay URL, and three 32-byte X25519 keys
  (local private + public, peer public). This is what makes a session
  end-to-end-encrypted and was the only persistence primitive before
  v0.19.0.
- **`PairingArtifact`** — a persistable+expirable envelope around a
  `PairingRecord`. Adds a bearer token, an issued-at timestamp, and an
  expires-at timestamp. The artifact is what gets transported (QR
  payload, xcrun env var, file in the app container) and what gets
  stored on the device.
- **`CredentialStore`** — a uniform interface (`Save` / `Load` /
  `Delete` / `IsExpired`) that a client uses to persist the artifact
  across launches. Reference implementations: file (Go/Swift/Kotlin),
  Keychain (Swift, iOS/macOS).
- **`PairingHost`** — server-side artifact minter. Wraps the same
  in-process keypair generation as `IssueCredential`, plus a TTL
  (default 30 days) and an optional bearer-token issuer.
- **Expiry routing** — `ConnectWithArtifact` (Go),
  `PigeonConn.connect(artifact:)` (Swift), `connectWithArtifact`
  (Kotlin) all check `IsExpired` up front and surface a typed error
  the application can branch on for re-pair.

## TTL — why 30 days

The default `DefaultPairingTTL` is 30 days. The choice is a balance
between two failure modes:

- **Too short**: legitimate users hit re-pair friction often. Re-pair
  is a multi-step ceremony — for an iOS app deployed via xcrun it
  means redeploying; for a QR-scan flow it means scanning again.
- **Too long**: a leaked artifact stays useful for the attacker for
  too long. (The relay never sees plaintext, and the application
  layer can revoke a server-side `PairingRecord` independently of TTL,
  but the artifact itself is what gets stolen, and TTL is the
  defence-in-depth bound.)

30 days places re-pair friction inside a typical "I'll dig out my
phone every few weeks" cadence. Override per-host if you have stronger
opinions; values below 1 day are usually a smell. Negative TTL
(`PairingHost.TTL = -1`) produces non-expiring artifacts — use only
in development.

## Two delivery flows

The artifact is transport-agnostic. Two flows are first-class:

### Flow 1 — QR scan (paired-screen UX)

The pairing initiator (the laptop, or a CLI on the same machine as
the server) renders the artifact's text encoding as a QR code. The
client device scans it. Suitable for end-user pairing where there's
no privileged access to the device.

### Flow 2 — Developer-deploy via xcrun

The deploy script mints the artifact on the laptop and injects it
into the app's launch environment via `xcrun` during install. Suitable
for development, beta builds where the operator has device-level
access, and CI pipelines.

Both flows produce the same artifact bytes; only the transport
differs. The library doesn't care which one you used, and a single
client can support both (e.g. xcrun for dev builds, QR for production).

---

## Lifecycle — step by step

### Step 1: Server mints an artifact

A server-side process produces a fresh artifact for a peer instance
ID. The companion server-side `PairingRecord` must be persisted by
the application before the client tries to reconnect — otherwise the
auth handshake will be rejected.

#### Go (in-process)

```go
host := pigeon.NewPairingHost("https://relay.example.com")
artifact, serverRecord, err := host.Mint("device-pippa")
if err != nil { ... }
// Persist serverRecord in your application's per-device credential map.
// Transport `artifact` to the client via QR or xcrun.
```

#### CLI (`pigeon pair`) — for deploy scripts

```bash
pigeon pair --relay=https://relay.example.com --instance=device-pippa --format=text
```

- Stdout: the artifact in canonical text encoding (base64url JSON, no
  padding).
- Stderr: the server-side `PairingRecord` JSON. Capture with
  `--server-record-out=path/to/server.json` or `2>server.json`.

### Step 2: Deliver the artifact

#### QR scan

```go
// Server-side: render the artifact as a QR payload.
text, _ := artifact.MarshalText()
qr.Print(os.Stdout, string(text))   // or rasterise as an image for a screen
```

The client scans, decodes via `ParsePairingArtifactText` /
`PairingArtifact.fromText` / `PairingArtifact.fromText`, and proceeds
to Step 3.

#### Developer-deploy via xcrun

A single-liner deploys the app and launches it with the artifact in
its environment:

```bash
xcrun devicectl device install app --device <UDID> MyApp.app && \
xcrun devicectl device process launch --device <UDID> com.example.MyApp \
  --environment-variables PIGEON_PAIRING_ARTIFACT="$(pigeon pair \
      --relay=https://relay.example.com --instance=device-pippa --format=text \
      2>/tmp/pigeon-server.json)"
```

The `2>/tmp/pigeon-server.json` redirect captures the server-side
record for your backend to load (see the **Server-side acceptance
pattern** section below).

### Step 3: Client persists the artifact

On first launch the app saves the artifact to a `CredentialStore`.
Subsequent launches load it from the store instead of looking at the
delivery channel.

#### Go

```go
store := pigeon.NewFileCredentialStore("/var/lib/myapp/pigeon-artifact.json")

if envText := os.Getenv("PIGEON_PAIRING_ARTIFACT"); envText != "" {
    a, err := pigeon.ParsePairingArtifactText([]byte(envText))
    if err != nil { ... }
    if err := store.Save(a); err != nil { ... }
}
artifact, err := store.Load()
if errors.Is(err, pigeon.ErrNoCredential) {
    // Prompt for re-pair / first-time pairing.
}
```

#### Swift

```swift
let store = KeychainCredentialStore(service: "MyApp")

if let envText = ProcessInfo.processInfo.environment["PIGEON_PAIRING_ARTIFACT"],
   !envText.isEmpty {
    let fresh = try PairingArtifact.fromText(envText)
    try store.save(fresh)
}
let artifact: PairingArtifact
do {
    artifact = try store.load()
} catch CredentialStoreError.noCredential {
    // Prompt for re-pair / first-time pairing.
    return
}
```

#### Kotlin

```kotlin
val store = FileCredentialStore(File(context.filesDir, "pigeon-artifact.json"))

System.getenv("PIGEON_PAIRING_ARTIFACT")?.takeIf { it.isNotEmpty() }?.let {
    store.save(PairingArtifact.fromText(it))
}
val artifact = try {
    store.load()
} catch (e: NoCredentialException) {
    // Prompt for re-pair / first-time pairing.
    return
}
```

### Step 4: Connect using the artifact

The artifact-driven `connect` helpers check expiry up front and dial
in one call. Expiry surfaces as a typed error so the application can
route to a re-pair flow uniformly.

#### Go

> **Note:** The `ConnectWithArtifact` helper (and the surrounding
> `PairingArtifact` / `CredentialStore` / `PairingHost` API) is not yet
> implemented in Go. Use `pigeon.Connect` with a `*crypto.PairingRecord`
> loaded from your own store. See the root `api.go` for the current API.

#### Swift

```swift
do {
    let (conn, channel) = try await PigeonConn.connect(artifact: artifact)
    // Swift's PigeonConn doesn't have a setChannel method — you wrap
    // E2EChannel.encrypt/decrypt around send/recv manually.
    let ciphertext = try channel.encrypt(Data("hello".utf8))
    try await conn.send(ciphertext)
    let response = try channel.decrypt(try await conn.recv())
} catch PairingError.expired {
    try store.delete()
    // route to re-pair UI
}
```

> **Swift API shape note.** Unlike Go and Kotlin (which return a
> `PigeonConn` with the channel already wired), Swift's
> `PigeonConn.connect(artifact:)` returns `(PigeonConn, E2EChannel)`.
> The Swift `PigeonConn` doesn't currently have a `setChannel`
> method, so the caller wraps `channel.encrypt` / `channel.decrypt`
> around `conn.send` / `conn.recv` themselves.

#### Kotlin

```kotlin
val transport = KwikQuicTransport.connect(host, port)  // your QUIC transport
val conn = try {
    connectWithArtifact(transport, artifact)
} catch (e: PairingExpiredException) {
    store.delete()
    // route to re-pair UI
    return
}
conn.send("hello".toByteArray())
```

### Step 5: Server-side acceptance

When the device reconnects, the server-side application must look up
the matching `PairingRecord` (saved at Step 1) and attach it to the
inbound connection so per-channel encryption derives correctly.

```go
// Your application's per-device PairingRecord map. Back it with
// whatever you already use (Postgres, BoltDB, file).
type CredentialStore struct {
    mu sync.RWMutex
    m  map[string]*crypto.PairingRecord
}

// Loading the stderr capture from `pigeon pair`:
data, err := os.ReadFile("/tmp/pigeon-server.json")
rec, err := crypto.UnmarshalPairingRecord(data)
store.Save(rec)
```

Inside your `pigeon.Register` accept loop, pass the record lookup as the `Pairing` function so activation resolves it per-client:

```go
listener, instanceID, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
    Identity: identity,
    Pairing: func(clientID string) (*crypto.PairingRecord, error) {
        rec, ok := store.Load(clientID)
        if !ok {
            return nil, fmt.Errorf("unknown device %q — re-pair required", clientID)
        }
        return rec, nil
    },
    Relay: relayURL,
    Token: bearerToken,
})
defer listener.Close()
// Each Accept call returns a fully-activated *Session with AEAD derived
// from the matching PairingRecord. No SetPairingRecord call needed.
session, err := listener.Accept(ctx)
```

### Step 6: Expiry triggers re-pair

When the artifact passes its `ExpiresAt`, the next call to
`ConnectWithArtifact` (or its Swift/Kotlin equivalent) returns
`ErrPairingExpired` / `PairingError.expired` /
`PairingExpiredException` *before any network IO*. The application
deletes the stale credential and routes to whichever re-pair flow
fits the deployment:

- **xcrun-deployed apps**: re-run the deploy single-liner. The new
  `PIGEON_PAIRING_ARTIFACT` env var is picked up by the bootstrap
  code at Step 3 and saved over the deleted credential.
- **QR-paired apps**: surface the QR scan UI again; the user re-runs
  the pairing ceremony.

Pigeon doesn't decide which flow you use — it surfaces the typed
error and steps out of the way.

---

## Security notes

- **The artifact carries `local_private_key`.** A client device that
  loses control of its artifact loses control of its identity for
  whatever's left of the TTL. The relay still can't read the
  device's traffic (it doesn't have the server-side `PairingRecord`),
  but an attacker with the artifact can impersonate the device until
  the artifact expires or the application revokes the server-side
  record.
- **Keychain on iOS, EncryptedSharedPreferences on Android, file with
  `0600` on desktop.** All three reference `CredentialStore`
  implementations apply the platform's strongest at-rest protection.
  Don't bypass the store and write the artifact to a normal file in a
  shared directory.
- **Don't log the artifact.** The text encoding is short (~500 bytes
  base64url) and looks innocuous; treat it like a password.
- **`pigeon pair` writes the server-side `PairingRecord` to stderr by
  default.** Capture it explicitly with `2>path/to/server.json` so it
  doesn't leak into terminal scrollback or CI logs. The server-side
  record contains the server's private key for that device — losing
  it is equivalent to losing the device's identity from the server's
  perspective.
- **`PIGEON_PAIRING_ARTIFACT` env var leakage.** If an attacker reads
  the env block of your app process (e.g. via `/proc/<pid>/environ`
  on a rooted device, or shell history on the laptop), they get the
  artifact. Save to the credential store on first launch and clear
  the env var if your runtime allows; on iOS the env var is scoped to
  the launched process and goes away when the app exits.

## Cross-SDK wire compatibility

The artifact's canonical encoding is the same on all three SDKs:

- JSON, snake_case keys, ISO-8601 timestamps, base64-encoded byte fields.
- Text encoding: base64url of the JSON bytes, no padding.

A Go-minted artifact decodes in Swift via `PairingArtifact.fromText`
or `PairingArtifact.fromJSON`; a Swift-minted artifact decodes in
Kotlin via `PairingArtifact.fromText` / `PairingArtifact.fromJson`.
The only thing that varies between SDKs is the API surface around
the artifact — the bytes are universal.

## Reference

- [`STABILITY.md`](../STABILITY.md) — full type signatures for all
  three SDKs.
- [`agents-guide.md`](../agents-guide.md) — top-level pigeon
  reference for coding agents.
- v0.19.0 release notes — the release that introduced this surface.
