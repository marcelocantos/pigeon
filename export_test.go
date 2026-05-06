// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

// MachineState exposes the current SessionMachine state for tests in
// the external pigeon_test package (e2e_test.go and friends). The
// machine itself stays unexported; tests should treat the state value
// as opaque except for equality comparisons against the
// SessionProtocol*State constants.
func (s *Session) MachineState() State {
	if s.machine == nil {
		return ""
	}
	return s.machine.State()
}
