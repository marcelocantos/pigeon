// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Command tern is a WebTransport relay server. Backend instances
// register and receive a unique instance ID. Clients connect by ID
// and all traffic is forwarded bidirectionally (streams and datagrams).
//
// Endpoints (served over HTTP/3):
//
//	GET /health             — health check
//	GET /register           — backend connects here (WebTransport session)
//	GET /ws/<instance-id>   — client connects here (WebTransport session)
package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/marcelocantos/tern"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0".
var version = "dev"

// generateSelfSignedCert creates a self-signed TLS certificate for
// development use. Production deployments should provide a real certificate.
func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	helpAgent := flag.Bool("help-agent", false, "print help and agent guide")
	port := flag.String("port", "", "listening port (overrides PORT env var)")
	certFile := flag.String("cert", "", "TLS certificate file (PEM)")
	keyFile := flag.String("key", "", "TLS private key file (PEM)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *helpAgent {
		var buf bytes.Buffer
		flag.CommandLine.SetOutput(&buf)
		flag.Usage()
		fmt.Print(buf.String())
		fmt.Println(tern.AgentGuide)
		os.Exit(0)
	}

	listenPort := *port
	if listenPort == "" {
		listenPort = os.Getenv("PORT")
	}
	if listenPort == "" {
		listenPort = "443"
	}

	// TERN_TOKEN restricts /register to authorized backends.
	// If unset, registration is open.
	token := os.Getenv("TERN_TOKEN")

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load or generate TLS certificate.
	var tlsCert tls.Certificate
	var err error
	if *certFile != "" && *keyFile != "" {
		tlsCert, err = tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			slog.Error("failed to load TLS certificate", "err", err)
			os.Exit(1)
		}
		slog.Info("loaded TLS certificate", "cert", *certFile)
	} else {
		tlsCert, err = generateSelfSignedCert()
		if err != nil {
			slog.Error("failed to generate self-signed certificate", "err", err)
			os.Exit(1)
		}
		slog.Info("generated self-signed TLS certificate (development mode)")
	}

	addr := ":" + listenPort
	srv, err := tern.NewWebTransportServer(addr, tlsCert, token)
	if err != nil {
		slog.Error("failed to create server", "err", err)
		os.Exit(1)
	}

	slog.Info("tern starting", "addr", addr, "version", version, "transport", "webtransport")
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("tern failed", "err", err)
		os.Exit(1)
	}
}
