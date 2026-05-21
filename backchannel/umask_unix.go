// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package backchannel

import (
	"sync"
	"syscall"
	"time"
)

// umaskMu serialises umask manipulation across the backchannel
// package. Umask is process-wide; concurrent Listen calls would
// otherwise leak a permissive mask to whatever goroutines happened to
// create files in between.
var umaskMu sync.Mutex

// withUmask runs fn with the process umask temporarily set to mask,
// restoring the prior value before returning. The umaskMu lock is held
// for the duration of fn so concurrent file/socket creations elsewhere
// in the program do not pick up our restrictive mask. fn should be
// short — the bind syscall only.
func withUmask(mask int, fn func() error) error {
	umaskMu.Lock()
	defer umaskMu.Unlock()
	prev := syscall.Umask(mask)
	defer syscall.Umask(prev)
	return fn()
}

// deadlineEpoch is a time value in the distant past, used to force any
// pending Accept on a net.UnixListener to return immediately. Using
// time.Unix(1, 0) is portable across Go versions; time.Time{} (zero
// value) means "no deadline" which would not interrupt the wait.
func deadlineEpoch() time.Time { return time.Unix(1, 0) }
