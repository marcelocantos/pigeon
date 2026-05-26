// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/x509"
	"errors"
	"testing"
)

func TestBearerTokenAuthAcceptsMatchingToken(t *testing.T) {
	a := BearerTokenAuth("s3cret")
	if a.VerifyRegister == nil {
		t.Fatal("VerifyRegister must be set for non-empty token")
	}
	if err := a.VerifyRegister(context.Background(), &RegisterRequest{Token: "s3cret"}); err != nil {
		t.Fatalf("matching token rejected: %v", err)
	}
}

func TestBearerTokenAuthRejectsMismatch(t *testing.T) {
	a := BearerTokenAuth("s3cret")
	err := a.VerifyRegister(context.Background(), &RegisterRequest{Token: "wrong"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	err = a.VerifyRegister(context.Background(), &RegisterRequest{Token: ""})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized for empty presented token, got %v", err)
	}
}

func TestBearerTokenAuthEmptyTokenIsOpen(t *testing.T) {
	a := BearerTokenAuth("")
	if a.VerifyRegister != nil || a.VerifyConnect != nil {
		t.Fatal("empty token must yield the zero Auth (open relay)")
	}
}

func TestMutualTLSAuthRejectsWithoutPeerCert(t *testing.T) {
	a := MutualTLSAuth(x509.NewCertPool())
	if err := a.VerifyRegister(context.Background(), &RegisterRequest{}); err == nil {
		t.Fatal("expected rejection when no peer cert presented")
	}
}
