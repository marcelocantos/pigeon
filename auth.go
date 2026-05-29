// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"

	"github.com/quic-go/quic-go"
)

// RegisterRequest is the verifier-visible state at the point where a
// backend's register greeting has been received: the QUIC (or
// WebTransport-over-QUIC) handshake has completed and the greeting has
// been parsed, but registration has not yet taken effect.
//
// Auth verifiers inspect this struct to decide whether to admit the
// backend. The Token, TLS, QUICConn and HTTPRequest fields expose
// orthogonal mechanisms — a verifier may consult one, several, or all
// of them depending on the deployment.
type RegisterRequest struct {
	// Token is the greeting-level credential, if any. For raw QUIC
	// it is the optional token field of register / register-mux. For
	// WebTransport it is the Bearer credential from the Authorization
	// header (or the ?token= query parameter). Empty if no token was
	// presented.
	Token string

	// InstanceID is the backend-requested instance ID. Empty means
	// "assign one".
	InstanceID string

	// TLS is the negotiated TLS state — peer certificates (if mTLS
	// is in effect), ALPN, SNI, cipher suite. Always non-nil for the
	// QUIC and WebTransport transports.
	TLS *tls.ConnectionState

	// QUICConn exposes the underlying *quic.Conn for raw-QUIC
	// registrations. Nil for WebTransport. Use this to read QUIC
	// transport parameters or migration state.
	QUICConn *quic.Conn

	// HTTPRequest exposes the WebTransport upgrade request for
	// WebTransport registrations. Nil for raw QUIC. Use this to read
	// arbitrary headers or query parameters.
	HTTPRequest *http.Request
}

// ConnectRequest mirrors RegisterRequest for the client side: a
// client has asked to connect to a registered instance, and the
// verifier decides whether to admit the connection. Clients are
// typically authenticated transitively by the backend's own pairing
// logic — VerifyConnect is unset by default — but adopters with
// stricter requirements can enforce additional checks here.
type ConnectRequest struct {
	InstanceID  string
	TLS         *tls.ConnectionState
	QUICConn    *quic.Conn
	HTTPRequest *http.Request
}

// Auth is the server-side authentication hook. Both fields are
// optional; the zero Auth means "accept all" (open relay).
//
// Verifiers receive the live transport handle so they can inspect
// whatever the deployment needs — peer certificates for mTLS,
// transport parameters, headers, or the greeting token. Returning a
// non-nil error rejects the registration / connection.
type Auth struct {
	VerifyRegister func(ctx context.Context, req *RegisterRequest) error
	VerifyConnect  func(ctx context.Context, req *ConnectRequest) error
}

// ErrUnauthorized is the canonical rejection error returned by the
// bundled verifiers. Custom verifiers may return any non-nil error;
// the server treats all non-nil returns as a refusal.
var ErrUnauthorized = errors.New("unauthorized")

// BearerTokenAuth returns an Auth that admits only registrations
// whose greeting token (or WebTransport Bearer credential) matches
// the configured token in constant time. An empty token returns the
// zero Auth (accept-all), preserving the open-relay default.
//
// Clients are not gated — VerifyConnect is left unset. (Set it
// explicitly if you need client-side admission control.)
func BearerTokenAuth(token string) Auth {
	if token == "" {
		return Auth{}
	}
	return Auth{
		VerifyRegister: func(_ context.Context, req *RegisterRequest) error {
			if subtle.ConstantTimeCompare([]byte(req.Token), []byte(token)) != 1 {
				return ErrUnauthorized
			}
			return nil
		},
	}
}

// MutualTLSAuth returns an Auth that requires the peer to have
// presented a client certificate signed by one of the roots in pool
// during the TLS handshake.
//
// The caller is responsible for configuring tls.Config.ClientAuth
// (e.g. tls.RequireAndVerifyClientCert) and tls.Config.ClientCAs on
// the server TLS config — without those, the TLS stack never asks
// for a client cert and this verifier always rejects.
func MutualTLSAuth(pool *x509.CertPool) Auth {
	return Auth{
		VerifyRegister: func(_ context.Context, req *RegisterRequest) error {
			if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
				return ErrUnauthorized
			}
			if _, err := req.TLS.PeerCertificates[0].Verify(x509.VerifyOptions{Roots: pool}); err != nil {
				return err
			}
			return nil
		},
	}
}
