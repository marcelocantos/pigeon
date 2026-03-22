// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package qr

import (
	"bytes"
	"testing"
)

func TestLanIP(t *testing.T) {
	ip := LanIP()
	if ip == "" {
		t.Fatal("LanIP returned empty string")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	Print(&buf, "https://example.com")
	if buf.Len() == 0 {
		t.Fatal("Print wrote nothing to the buffer")
	}
}

func TestPrintInvalidURL(t *testing.T) {
	// Verify Print does not panic on empty string.
	var buf bytes.Buffer
	Print(&buf, "")
	// No assertion on content; just verifying it doesn't panic.
}
