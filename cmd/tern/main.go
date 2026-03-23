// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Command tern is a relay server supporting both WebTransport (for
// browsers) and raw QUIC (for native clients). Backend instances
// register and receive a unique instance ID. Clients connect by ID
// and all traffic is forwarded bidirectionally (streams and datagrams).
//
// Endpoints (WebTransport, HTTP/3):
//
//	GET /health             — health check
//	GET /register           — backend connects here (WebTransport session)
//	GET /ws/<instance-id>   — client connects here (WebTransport session)
//
// Raw QUIC (ALPN "tern"):
//
//	Handshake "register" or "register:<token>" — backend registration
//	Handshake "connect:<instance-id>"          — client connection
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
	"net/http"
	"os"
	"time"

	"github.com/caddyserver/certmagic"

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
	port := flag.String("port", "", "WebTransport listening port (overrides PORT env var)")
	quicPort := flag.String("quic-port", "", "raw QUIC listening port (overrides QUIC_PORT env var)")
	certFile := flag.String("cert", "", "TLS certificate file (PEM)")
	keyFile := flag.String("key", "", "TLS private key file (PEM)")
	domain := flag.String("domain", "", "domain for automatic Let's Encrypt TLS (e.g. tern.fly.dev)")
	acmeEmail := flag.String("acme-email", "", "email for Let's Encrypt account (recommended)")
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

	listenQUICPort := *quicPort
	if listenQUICPort == "" {
		listenQUICPort = os.Getenv("QUIC_PORT")
	}
	if listenQUICPort == "" {
		listenQUICPort = "4433"
	}

	// TERN_TOKEN restricts /register to authorized backends.
	// If unset, registration is open.
	token := os.Getenv("TERN_TOKEN")

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Build TLS config from one of three sources:
	// 1. --domain: automatic Let's Encrypt via certmagic
	// 2. --cert/--key: static PEM certificate files
	// 3. Neither: self-signed certificate (development mode)
	var tlsConfig *tls.Config

	switch {
	case *domain != "":
		// Configure certmagic for automatic Let's Encrypt certificates.
		certmagic.DefaultACME.Agreed = true
		if *acmeEmail != "" {
			certmagic.DefaultACME.Email = *acmeEmail
		}

		cfg := certmagic.NewDefault()
		if err := cfg.ManageSync(nil, []string{*domain}); err != nil {
			slog.Error("failed to provision certificate", "domain", *domain, "err", err)
			os.Exit(1)
		}

		tlsConfig = cfg.TLSConfig()
		slog.Info("using Let's Encrypt certificate", "domain", *domain)

	case *certFile != "" && *keyFile != "":
		tlsCert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			slog.Error("failed to load TLS certificate", "err", err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}
		slog.Info("loaded TLS certificate", "cert", *certFile)

	default:
		tlsCert, err := generateSelfSignedCert()
		if err != nil {
			slog.Error("failed to generate self-signed certificate", "err", err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}
		slog.Info("generated self-signed TLS certificate (development mode)")
	}

	wtAddr := ":" + listenPort
	srv, err := tern.NewWebTransportServer(wtAddr, tlsConfig, token)
	if err != nil {
		slog.Error("failed to create WebTransport server", "err", err)
		os.Exit(1)
	}

	// Start the raw QUIC server sharing the same hub.
	qAddr := ":" + listenQUICPort
	qsrv := tern.NewQUICServer(qAddr, tlsConfig, token, srv.Hub())

	// When using certmagic, start a TCP/TLS listener on the same port
	// for ACME TLS-ALPN-01 challenges and HTTPS health checks.
	if *domain != "" {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Alt-Svc header tells browsers that HTTP/3 (and thus WebTransport)
			// is available on the same port. Without this, browsers won't attempt
			// QUIC and WebTransport handshakes will fail.
			w.Header().Set("Alt-Svc", `h3=":443"; ma=86400`)

			if r.URL.Path == "/health" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"ok"}`))
				return
			}
			http.NotFound(w, r)
		})

		tcpTLS := tlsConfig.Clone()
		tcpTLS.NextProtos = []string{"h2", "http/1.1", "acme-tls/1"}

		tcpListener, err := tls.Listen("tcp", wtAddr, tcpTLS)
		if err != nil {
			slog.Error("failed to start TCP/TLS listener", "err", err)
			os.Exit(1)
		}
		go func() {
			slog.Info("HTTPS listener started", "addr", wtAddr)
			if err := http.Serve(tcpListener, healthMux); err != nil {
				slog.Error("HTTPS listener failed", "err", err)
			}
		}()
	}

	// Start raw QUIC server in background.
	go func() {
		slog.Info("raw QUIC server starting", "addr", qAddr)
		if err := qsrv.ListenAndServe(tlsConfig); err != nil {
			slog.Error("raw QUIC server failed", "err", err)
		}
	}()

	slog.Info("tern starting",
		"wt-addr", wtAddr,
		"quic-addr", qAddr,
		"version", version,
	)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("tern failed", "err", err)
		os.Exit(1)
	}
}
