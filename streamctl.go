// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

// streamCtl is the optional FIN/RST surface on a QUIC-backed stream.
// CloseWrite emits STREAM FIN (write half-close). Reset emits
// RESET_STREAM + STOP_SENDING. Implementations that only have Close()
// treat both as Close.
type streamCtl interface {
	CloseWrite() error
	Reset() error
}

func closeWrite(rwc io.ReadWriteCloser) error {
	if rwc == nil {
		return nil
	}
	if c, ok := rwc.(streamCtl); ok {
		return c.CloseWrite()
	}
	switch s := rwc.(type) {
	case *quic.Stream:
		return s.Close()
	case *webtransport.Stream:
		return s.Close()
	default:
		return rwc.Close()
	}
}

func resetStream(rwc io.ReadWriteCloser) error {
	if rwc == nil {
		return nil
	}
	if c, ok := rwc.(streamCtl); ok {
		return c.Reset()
	}
	switch s := rwc.(type) {
	case *quic.Stream:
		s.CancelWrite(0)
		s.CancelRead(0)
		return nil
	case *webtransport.Stream:
		s.CancelWrite(0)
		s.CancelRead(0)
		return nil
	default:
		return rwc.Close()
	}
}

func isStreamReset(err error) bool {
	if err == nil {
		return false
	}
	var qe *quic.StreamError
	if errors.As(err, &qe) {
		return true
	}
	var we *webtransport.StreamError
	return errors.As(err, &we)
}

func closeWritePair(p readWriteCloserPair) {
	if p == nil {
		return
	}
	if c, ok := p.(streamCtl); ok {
		_ = c.CloseWrite()
		return
	}
	_ = p.Close()
}

func resetPair(p readWriteCloserPair) {
	if p == nil {
		return
	}
	if c, ok := p.(streamCtl); ok {
		_ = c.Reset()
		return
	}
	_ = p.Close()
}
