// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package httpguard rejects requests that this relay should never have
// been asked to serve.
//
// Binding a port is not, on its own, a security boundary. A web page
// can point a hostname it controls at the relay (DNS rebinding) and
// then have the browser issue same-origin requests. The browser still
// sends the attacker's name in Host, and its own page origin in
// Origin. Checking both is the defence.
//
// The relay is a public rendezvous, so a browser Origin from any
// http(s) app is accepted. null and non-http(s) origins are not.
// Host must name this relay: a loopback literal, a unicast IP, or
// the configured --domain.
package httpguard

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	localhostName   = "localhost"
	defaultHTTPPort = "80"
	defaultTLSPort  = "443"
	nullOrigin      = "null"
)

// AllowRequest is the Host/Origin guard for the WebTransport upgrade
// and the HTTPS diagnostic mux.
func AllowRequest(r *http.Request, boundPort, domain string) bool {
	return allowHost(r.Host, boundPort, domain) && allowOrigin(r.Header.Get("Origin"))
}

func allowOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	if origin == nullOrigin {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	return true
}

func allowHost(hostHeader, boundPort, domain string) bool {
	host, hostPort := splitHostPort(hostHeader)
	if !hostNameOK(host, domain) {
		return false
	}
	if hostPort == "" || hostPort == boundPort {
		return true
	}
	if boundPort == defaultTLSPort && hostPort == defaultHTTPPort {
		// A Host header with no port is reported as 80 by the splitter;
		// TLS relays still accept that form of the configured name.
		return hostHeader != "" && !strings.Contains(hostHeader, ":")
	}
	return hostPort == boundPort
}

func hostNameOK(host, domain string) bool {
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return false
	}
	if strings.EqualFold(host, localhostName) {
		return true
	}
	if domain != "" && strings.EqualFold(host, strings.TrimSuffix(domain, ".")) {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLinkLocalUnicast())
}

// PortFromListenAddr returns the port in an addr such as ":443" or "127.0.0.1:4433".
func PortFromListenAddr(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil {
		return p
	}
	return strings.TrimPrefix(addr, ":")
}

func splitHostPort(hostport string) (host, port string) {
	if hostport == "" {
		return "", defaultHTTPPort
	}
	if h, p, err := net.SplitHostPort(hostport); err == nil {
		return h, p
	}
	return strings.Trim(hostport, "[]"), defaultHTTPPort
}

// WriteDiagCORS sets CORS headers on /health and /status for browser
// Alt-Svc priming. Reflects a permitted Origin; never writes `*`.
func WriteDiagCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" || !allowOrigin(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Vary", "Origin")
}
