// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Command pigeon-pair mints a PairingArtifact for a peer instance and
// emits it on stdout. Suitable for deploy scripts that need to inject
// a credential onto a device out-of-band — for example, an iOS
// pipeline that runs:
//
//	pigeon-pair --relay=https://relay.example.com --instance=device-42 \
//	    --ttl=30d --format=text > artifact.txt
//	xcrun devicectl device copy to ... artifact.txt /var/.../Documents/
//	xcrun devicectl device process launch <bundle>
//
// The companion server-side PairingRecord is emitted on stderr in JSON
// form so the deploy script can register it with the relay's auth
// sub-machine; redirect stderr or pipe through `tee` to capture it.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marcelocantos/pigeon"
)

func main() {
	relay := flag.String("relay", "", "relay URL embedded in the artifact (required)")
	instance := flag.String("instance", "", "peer instance ID the device will register with (required)")
	ttl := flag.Duration("ttl", pigeon.DefaultPairingTTL, "artifact lifetime (e.g. 720h, 30d-equivalent); -1 for no expiry")
	token := flag.String("token", "", "bearer token to embed in the artifact (optional)")
	format := flag.String("format", "json", "artifact output format: \"json\" or \"text\" (base64url)")
	out := flag.String("out", "-", "artifact output path; \"-\" for stdout")
	serverOut := flag.String("server-record-out", "-stderr-", "server-side PairingRecord output path; \"-stderr-\" for stderr, \"-\" for stdout, or a file path")
	flag.Parse()

	if *relay == "" || *instance == "" {
		fmt.Fprintln(os.Stderr, "pigeon-pair: --relay and --instance are required")
		flag.Usage()
		os.Exit(2)
	}

	host := &pigeon.PairingHost{
		RelayURL: *relay,
		TTL:      *ttl,
	}
	if *token != "" {
		host.IssueToken = func(string) (string, error) { return *token, nil }
	}

	artifact, serverRec, err := host.Mint(*instance)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon-pair: mint:", err)
		os.Exit(1)
	}

	artifactBytes, err := encodeArtifact(artifact, *format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon-pair: encode:", err)
		os.Exit(1)
	}
	if err := writeOutput(*out, artifactBytes); err != nil {
		fmt.Fprintln(os.Stderr, "pigeon-pair: write artifact:", err)
		os.Exit(1)
	}

	serverBytes, err := serverRec.Marshal()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pigeon-pair: marshal server record:", err)
		os.Exit(1)
	}
	serverBytes = append(serverBytes, '\n')
	if err := writeServerRecord(*serverOut, serverBytes); err != nil {
		fmt.Fprintln(os.Stderr, "pigeon-pair: write server record:", err)
		os.Exit(1)
	}

	if *out != "-" {
		fmt.Fprintf(os.Stderr, "pigeon-pair: artifact for %s expires %s\n",
			*instance, formatExpiry(artifact))
	}
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

func writeOutput(path string, data []byte) error {
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
