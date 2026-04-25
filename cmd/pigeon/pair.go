// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marcelocantos/pigeon"
)

// runPair handles `pigeon pair ...`. It mints a PairingArtifact for a
// peer instance and emits it on stdout (or a file). Suitable for
// deploy scripts that need to inject a credential onto a device
// out-of-band — for example:
//
//	pigeon pair --relay=https://relay.example.com --instance=device-42 \
//	            --format=text |
//	xcrun devicectl device pasteboard set --device <UDID> -
//
// The companion server-side PairingRecord is emitted on stderr in JSON
// form so the deploy script can register it with the relay's auth
// sub-machine.
func runPair(args []string) int {
	fs := flag.NewFlagSet("pigeon pair", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	relay := fs.String("relay", "", "relay URL embedded in the artifact (required)")
	instance := fs.String("instance", "", "peer instance ID the device will register with (required)")
	ttl := fs.Duration("ttl", pigeon.DefaultPairingTTL, "artifact lifetime (e.g. 720h); -1 for no expiry")
	token := fs.String("token", "", "bearer token to embed in the artifact (optional)")
	format := fs.String("format", "json", "artifact output format: \"json\" or \"text\" (base64url)")
	out := fs.String("out", "-", "artifact output path; \"-\" for stdout")
	serverOut := fs.String("server-record-out", "-stderr-",
		"server-side PairingRecord output path; \"-stderr-\" for stderr, \"-\" for stdout, or a file path")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *relay == "" || *instance == "" {
		fmt.Fprintln(os.Stderr, "pigeon pair: --relay and --instance are required")
		fs.Usage()
		return 2
	}

	host := &pigeon.PairingHost{RelayURL: *relay, TTL: *ttl}
	if *token != "" {
		host.IssueToken = func(string) (string, error) { return *token, nil }
	}

	artifact, serverRec, err := host.Mint(*instance)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon pair: mint:", err)
		return 1
	}

	artifactBytes, err := encodeArtifact(artifact, *format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon pair: encode:", err)
		return 1
	}
	if err := writePath(*out, artifactBytes); err != nil {
		fmt.Fprintln(os.Stderr, "pigeon pair: write artifact:", err)
		return 1
	}

	serverBytes, err := serverRec.Marshal()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon pair: marshal server record:", err)
		return 1
	}
	serverBytes = append(serverBytes, '\n')
	if err := writeServerRecord(*serverOut, serverBytes); err != nil {
		fmt.Fprintln(os.Stderr, "pigeon pair: write server record:", err)
		return 1
	}

	if *out != "-" {
		fmt.Fprintf(os.Stderr, "pigeon pair: artifact for %s expires %s\n",
			*instance, formatExpiry(artifact))
	}
	return 0
}

func encodeArtifact(a *pigeon.PairingArtifact, format string) ([]byte, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(data, '\n'), nil
	case "text":
		data, err := a.MarshalText()
		if err != nil {
			return nil, err
		}
		return append(data, '\n'), nil
	default:
		return nil, fmt.Errorf("unknown format %q (want \"json\" or \"text\")", format)
	}
}

func writePath(path string, data []byte) error {
	if path == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func writeServerRecord(path string, data []byte) error {
	switch path {
	case "-":
		_, err := os.Stdout.Write(data)
		return err
	case "-stderr-":
		_, err := os.Stderr.Write(data)
		return err
	default:
		return os.WriteFile(path, data, 0o600)
	}
}

func formatExpiry(a *pigeon.PairingArtifact) string {
	if a.ExpiresAt.IsZero() {
		return "never"
	}
	return a.ExpiresAt.UTC().Format(time.RFC3339)
}
