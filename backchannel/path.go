// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SocketPath returns the well-known Unix-socket path for appName on the
// current host. Linux uses $XDG_RUNTIME_DIR/pigeon/<app>.sock (falling
// back to /tmp when the env var is unset, as per the XDG spec for
// non-session contexts); macOS uses ~/Library/Application
// Support/pigeon/<app>.sock.
//
// appName must contain only [a-z0-9._-] so it is safe to use as a
// filename component on every supported filesystem.
func SocketPath(appName string) (string, error) {
	if err := validateAppName(appName); err != nil {
		return "", err
	}
	dir, err := socketDir(appName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName+".sock"), nil
}

// socketDir returns the parent directory the socket lives in. The
// daemon creates this with mode 0700.
func socketDir(appName string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("backchannel: locate $HOME: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "pigeon"), nil
	case "linux":
		// $XDG_RUNTIME_DIR is the canonical location for per-user
		// non-essential runtime sockets. When unset (e.g. headless
		// containers, systemd-less hosts) we degrade to /tmp/pigeon-$UID
		// which is owned by the calling user with mode 0700.
		if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
			return filepath.Join(xdg, "pigeon"), nil
		}
		return filepath.Join(os.TempDir(), fmt.Sprintf("pigeon-%d", os.Getuid())), nil
	default:
		return "", fmt.Errorf("backchannel: GOOS %q not supported (Windows tracked separately)", runtime.GOOS)
	}
}

func validateAppName(appName string) error {
	if appName == "" {
		return errors.New("backchannel: AppName is required")
	}
	for _, r := range appName {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '_' || r == '-':
		default:
			return fmt.Errorf("backchannel: AppName %q contains disallowed character %q (allow a-z 0-9 . _ -)", appName, r)
		}
	}
	if strings.HasPrefix(appName, ".") {
		return fmt.Errorf("backchannel: AppName %q must not start with .", appName)
	}
	return nil
}
