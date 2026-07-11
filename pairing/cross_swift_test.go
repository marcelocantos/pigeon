// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pairing

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/cwire"
)

// TestCrossLanguagePairingGoAcceptorSwiftInitiator runs the pairing
// ceremony with Go as the acceptor (via cwire.RunAcceptor / libpigeon's
// pigeon_pair_acceptor) and Swift as the initiator (via pairInitiator
// in Sources/Pigeon/PairingCeremony.swift, run from the
// `pairing-peer-swift` executable).
//
// The two halves communicate over the Swift subprocess's stdin/stdout
// using Go pairing.singleStreamTransport's wire format (4-byte BE length
// prefix per message). Acceptance per 🎯T25:
//   - both sides reach Paired,
//   - both sides derive identical 6-digit confirmation codes,
//   - the resulting PairingRecords' DeriveChannel outputs interoperate
//     (Go-encrypted ciphertext decrypts on Swift and vice-versa — the
//     latter via re-deriving the Swift channel with CryptoKit and
//     decrypting Go's output).
//
// The Swift binary is built on-demand if missing. Skipped on non-darwin
// since the Swift package only ships macOS toolchain builds today.
func TestCrossLanguagePairingGoAcceptorSwiftInitiator(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "darwin" {
		t.Skip("cross-language Swift test requires macOS toolchain")
	}

	// Locate repo root: ascend from test source dir until we see go.mod.
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}

	// Build the Swift initiator binary. swift build is incremental, so
	// repeated invocations are fast (~1s) once the package is hot.
	buildCtx, buildCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer buildCancel()
	build := exec.CommandContext(buildCtx, "swift", "build", "--product", "pairing-peer-swift")
	build.Dir = repoRoot
	var buildErr bytes.Buffer
	build.Stderr = &buildErr
	if err := build.Run(); err != nil {
		t.Fatalf("swift build pairing-peer-swift: %v\n%s", err, buildErr.String())
	}

	// Resolve the built binary path. swift build deposits products
	// under .build/debug/<product-name> by default.
	binPath := filepath.Join(repoRoot, ".build", "debug", "pairing-peer-swift")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("pairing-peer-swift not at %s: %v", binPath, err)
	}

	// Generate fresh keypairs on the Go (acceptor) side.
	accEph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("acceptor ephemeral: %v", err)
	}
	accIdKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("acceptor identity: %v", err)
	}
	const acceptorInstance = "go-acceptor-instance"

	// Generate fresh keypairs on the Swift (initiator) side via Go,
	// then base64-encode them to feed into the Swift binary's CLI flags.
	// Going through Go here is just a transport detail — the Swift
	// process will run the X25519 ECDH itself, the bytes just need to
	// be valid X25519 private/public material.
	initEph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("initiator ephemeral: %v", err)
	}
	initIdKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("initiator identity: %v", err)
	}
	const initiatorInstance = "swift-initiator-instance"

	// Launch the Swift initiator.
	stdB64 := base64.StdEncoding
	swiftCtx, swiftCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer swiftCancel()
	cmd := exec.CommandContext(swiftCtx, binPath,
		"--acceptor-eph-pub-b64="+stdB64.EncodeToString(accEph.PublicKey().Bytes()),
		"--acceptor-instance="+acceptorInstance,
		"--initiator-instance="+initiatorInstance,
		"--initiator-identity-pub-b64="+stdB64.EncodeToString(initIdKP.Public.Bytes()),
		"--initiator-eph-priv-b64="+stdB64.EncodeToString(initEph.Bytes()),
		"--initiator-eph-pub-b64="+stdB64.EncodeToString(initEph.PublicKey().Bytes()),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	var stderrBuf safeBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("start pairing-peer-swift: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	// Wrap subprocess stdio as an io.ReadWriteCloser; the singleStreamTransport
	// in cgo_pair.go reads 4-byte BE-len framed messages off it. Swift writes
	// hello/confirm to its stdout (= our `stdout`), and reads welcome/confirm
	// from its stdin (= our `stdin`).
	swiftRWC := &subprocessRWC{r: stdout, w: stdin}

	// Drive the Go acceptor via cwire.RunAcceptor.
	ref := cwire.NewGoTransportRef(&singleStreamTransport{rwc: swiftRWC})
	defer ref.Close()

	var goCode string
	rec, codeFromCwire, err := cwire.RunAcceptor(&cwire.RunAcceptorArgs{
		Ref:          ref,
		LocalEphPriv: accEph.Bytes(),
		LocalEphPub:  accEph.PublicKey().Bytes(),
		IdentityPub:  accIdKP.Public.Bytes(),
		InstanceID:   acceptorInstance,
		ConfirmFn: func(code string) bool {
			goCode = code
			return true
		},
	})
	if err != nil {
		// Surface the Swift binary's stderr to make Go-side failures
		// debuggable. Wait briefly for any trailing log lines.
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		t.Fatalf("cwire.RunAcceptor: %v\nswift stderr:\n%s", err, stderrBuf.String())
	}
	if codeFromCwire != goCode {
		t.Fatalf("codeFromCwire (%q) != confirmFn code (%q)", codeFromCwire, goCode)
	}

	// Wait for the Swift binary to finish and parse its RESULT line
	// from stderr — we expect a single "RESULT {json}\n" line.
	if err := cmd.Wait(); err != nil {
		t.Fatalf("pairing-peer-swift exit: %v\nstderr:\n%s", err, stderrBuf.String())
	}

	swiftResult := parseSwiftResult(t, stderrBuf.String())
	if swiftResult.Code != goCode {
		t.Fatalf("confirmation code mismatch: go=%q swift=%q\nswift stderr:\n%s",
			goCode, swiftResult.Code, stderrBuf.String())
	}
	if len(goCode) != 6 {
		t.Fatalf("code has wrong length: %q", goCode)
	}

	// Check that each side's PairingRecord references the other.
	if rec.PeerInstanceID != initiatorInstance {
		t.Errorf("go record peer = %q, want %q", rec.PeerInstanceID, initiatorInstance)
	}
	if swiftResult.Record.PeerInstanceID != acceptorInstance {
		t.Errorf("swift record peer = %q, want %q", swiftResult.Record.PeerInstanceID, acceptorInstance)
	}

	// The keys recorded by each side must agree with the other side's
	// material. The Swift record holds the initiator's keys, the Go
	// record holds the acceptor's; their peer_public_key fields cross-
	// reference.
	if !bytes.Equal(rec.LocalPubKey, accEph.PublicKey().Bytes()) {
		t.Errorf("go record local pub != acceptor ephemeral pub")
	}
	if !bytes.Equal(rec.PeerPubKey, initEph.PublicKey().Bytes()) {
		t.Errorf("go record peer pub != initiator ephemeral pub")
	}
	if !bytes.Equal(swiftResult.Record.LocalPublicKey, initEph.PublicKey().Bytes()) {
		t.Errorf("swift record local pub != initiator ephemeral pub")
	}
	if !bytes.Equal(swiftResult.Record.PeerPublicKey, accEph.PublicKey().Bytes()) {
		t.Errorf("swift record peer pub != acceptor ephemeral pub")
	}

	// Cross-DeriveChannel interop: build channels from each side's record
	// and round-trip a payload Go→Swift and Swift→Go.
	goRec := &crypto.PairingRecord{
		PeerInstanceID:  rec.PeerInstanceID,
		RelayURL:        "https://relay.test",
		LocalPrivateKey: rec.LocalPrivKey,
		LocalPublicKey:  rec.LocalPubKey,
		PeerPublicKey:   rec.PeerPubKey,
	}
	swiftRec := &crypto.PairingRecord{
		PeerInstanceID:  swiftResult.Record.PeerInstanceID,
		RelayURL:        swiftResult.Record.RelayURL,
		LocalPrivateKey: swiftResult.Record.LocalPrivateKey,
		LocalPublicKey:  swiftResult.Record.LocalPublicKey,
		PeerPublicKey:   swiftResult.Record.PeerPublicKey,
	}

	// Go channel: encrypts with "go->swift", decrypts with "swift->go".
	// Swift channel (re-derived on the Go side using its raw key material):
	// the opposite. Same ECDH shared secret produces matching session keys.
	goChan, err := goRec.DeriveChannel([]byte("go->swift"), []byte("swift->go"), nil)
	if err != nil {
		t.Fatalf("go DeriveChannel: %v", err)
	}
	swiftChan, err := swiftRec.DeriveChannel([]byte("swift->go"), []byte("go->swift"), nil)
	if err != nil {
		t.Fatalf("swift DeriveChannel: %v", err)
	}

	pt := []byte("hello from go acceptor over a cross-language channel")
	ct := goChan.Encrypt(pt)
	got, err := swiftChan.Decrypt(ct)
	if err != nil {
		t.Fatalf("swift decrypt of go ciphertext: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("go→swift payload mismatch: got %q want %q", got, pt)
	}

	pt2 := []byte("reply from swift initiator over the same channel")
	ct2 := swiftChan.Encrypt(pt2)
	got2, err := goChan.Decrypt(ct2)
	if err != nil {
		t.Fatalf("go decrypt of swift ciphertext: %v", err)
	}
	if !bytes.Equal(got2, pt2) {
		t.Fatalf("swift→go payload mismatch: got %q want %q", got2, pt2)
	}
}

// swiftPairingResult mirrors the JSON the Swift binary emits to stderr
// on the "RESULT " line after a successful ceremony.
type swiftPairingResult struct {
	Code   string `json:"code"`
	Record struct {
		PeerInstanceID  string `json:"peer_instance_id"`
		RelayURL        string `json:"relay_url"`
		LocalPrivateKey []byte `json:"local_private_key"`
		LocalPublicKey  []byte `json:"local_public_key"`
		PeerPublicKey   []byte `json:"peer_public_key"`
	} `json:"record"`
}

func parseSwiftResult(t *testing.T, stderr string) swiftPairingResult {
	t.Helper()
	scanner := bufio.NewScanner(strings.NewReader(stderr))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "RESULT ") {
			continue
		}
		jsonBlob := strings.TrimPrefix(line, "RESULT ")
		var res swiftPairingResult
		if err := json.Unmarshal([]byte(jsonBlob), &res); err != nil {
			t.Fatalf("parse RESULT line: %v\nline: %s", err, line)
		}
		return res
	}
	t.Fatalf("no RESULT line in swift stderr:\n%s", stderr)
	return swiftPairingResult{}
}

// subprocessRWC wraps an os.exec subprocess's stdio pipes into an
// io.ReadWriteCloser. Close shuts down stdin so the child sees EOF on
// its next read.
type subprocessRWC struct {
	r      io.Reader
	w      io.WriteCloser
	closed sync.Once
}

func (s *subprocessRWC) Read(p []byte) (int, error)  { return s.r.Read(p) }
func (s *subprocessRWC) Write(p []byte) (int, error) { return s.w.Write(p) }
func (s *subprocessRWC) Close() error {
	var err error
	s.closed.Do(func() { err = s.w.Close() })
	return err
}

// safeBuf is a tiny mutex-guarded bytes.Buffer so that os/exec's
// background goroutine writing the child's stderr doesn't race with the
// test goroutine reading it via String(). Without this, `go test -race`
// flags the Wait()-then-read sequence even though it's correct in
// practice (Wait blocks until the goroutine finishes).
type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}
func (b *safeBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// findRepoRoot walks up from this source file's directory until it
// finds a go.mod (the repo root).
func findRepoRoot() (string, error) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	dir := filepath.Dir(here)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod above %s", filepath.Dir(here))
		}
		dir = parent
	}
}
