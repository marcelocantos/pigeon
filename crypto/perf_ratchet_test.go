package crypto

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Allocation ratchet for the channel hot path.
//
// Wall time is not gated: this machine's throughput numbers move by tens
// of percent with whatever else is compiling. Allocation counts and bytes
// for a fixed payload do not move at all, and they are what regressed
// silently before — Encrypt allocated the ciphertext twice and copied it,
// costing 2.25 bytes of heap per byte relayed, and no test noticed.
//
// The lock holds in both directions: a regression fails, and so does an
// improvement, until it is re-recorded in the same commit with
//
//	PIGEON_PERF_RECORD=1 go test ./crypto/ -run Ratchet

const (
	recordEnv     = "PIGEON_PERF_RECORD"
	baselineFile  = "docs/perf/baseline.md"
	beginMarker   = "<!-- perf-baseline:begin -->"
	endMarker     = "<!-- perf-baseline:end -->"
	tolerancePct  = 1.0
	ratchetRounds = 5
)

type metric struct {
	allocs uint64
	bytes  uint64
}

// measure runs fn once untimed, then ratchetRounds times, and returns the
// minimum allocations and bytes observed. The minimum, not the mean:
// counts for a fixed input are already deterministic, and the minimum
// discards any accounting that a concurrent GC happens to attribute here.
func measure(fn func()) metric {
	fn()
	best := metric{^uint64(0), ^uint64(0)}
	for i := 0; i < ratchetRounds; i++ {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		fn()
		runtime.ReadMemStats(&after)
		if a := after.Mallocs - before.Mallocs; a < best.allocs {
			best.allocs = a
		}
		if b := after.TotalAlloc - before.TotalAlloc; b < best.bytes {
			best.bytes = b
		}
	}
	return best
}

func ratchetWorkloads(t *testing.T) map[string]func() {
	t.Helper()
	w := map[string]func(){}
	for _, n := range benchSizes {
		payload := make([]byte, n)
		for i := range payload {
			payload[i] = byte(i)
		}
		enc := benchChannel(t)
		w[fmt.Sprintf("encrypt/%d", n)] = func() { _ = enc.Encrypt(payload) }

		// Decrypt refuses a replayed sequence number, so every call needs
		// its own message. They are sealed up front, outside the measured
		// region, one per call measure will make.
		sender := benchChannel(t)
		msgs := make([][]byte, ratchetRounds+1)
		for i := range msgs {
			msgs[i] = sender.Encrypt(payload)
		}
		receiver, err := NewChannel(
			bytesOf(func(i int) byte { return byte(255 - i) }),
			bytesOf(func(i int) byte { return byte(i) }))
		if err != nil {
			t.Fatal(err)
		}
		next := 0
		w[fmt.Sprintf("decrypt/%d", n)] = func() {
			if next >= len(msgs) {
				t.Fatalf("decrypt workload ran %d times, only %d messages sealed",
					next+1, len(msgs))
			}
			if _, err := receiver.Decrypt(msgs[next]); err != nil {
				t.Fatal(err)
			}
			next++
		}
	}
	return w
}

func TestRatchetChannelAllocations(t *testing.T) {
	workloads := ratchetWorkloads(t)
	names := make([]string, 0, len(workloads))
	for name := range workloads {
		names = append(names, name)
	}
	sort.Strings(names)

	got := map[string]metric{}
	for _, name := range names {
		got[name] = measure(workloads[name])
	}

	path := baselinePath(t)
	if os.Getenv(recordEnv) == "1" {
		writeBaseline(t, path, names, got)
		t.Logf("recorded %d workloads to %s", len(names), baselineFile)
		return
	}

	want := readBaseline(t, path)
	for _, name := range names {
		w, ok := want[name]
		if !ok {
			t.Errorf("%s: no baseline row; record one with %s=1", name, recordEnv)
			continue
		}
		check(t, name, "allocs", got[name].allocs, w.allocs)
		check(t, name, "bytes", got[name].bytes, w.bytes)
	}
	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("%s: in baseline but no longer measured", name)
		}
	}
}

func check(t *testing.T, name, field string, got, want uint64) {
	t.Helper()
	if want == 0 {
		if got != 0 {
			t.Errorf("%s %s: 0 -> %d", name, field, got)
		}
		return
	}
	delta := 100 * (float64(got) - float64(want)) / float64(want)
	if delta > tolerancePct {
		t.Errorf("%s %s regressed: %d -> %d (%+.1f%%, band ±%.0f%%)",
			name, field, want, got, delta, tolerancePct)
	} else if -delta > tolerancePct {
		t.Errorf("%s %s improved: %d -> %d (%+.1f%%); lock it in with %s=1",
			name, field, want, got, delta, recordEnv)
	}
}

func baselinePath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, filepath.FromSlash(baselineFile))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("module root not found above %s", dir)
		}
		dir = parent
	}
}

func readBaseline(t *testing.T, path string) map[string]metric {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v (record one with %s=1)", baselineFile, err, recordEnv)
	}
	text := string(data)
	begin := strings.Index(text, beginMarker)
	end := strings.Index(text, endMarker)
	if begin < 0 || end < 0 {
		t.Fatalf("%s: missing %s block; record one with %s=1", baselineFile, beginMarker, recordEnv)
	}
	out := map[string]metric{}
	for _, line := range strings.Split(text[begin:end], "\n") {
		cells := splitRow(line)
		if len(cells) != 3 || cells[0] == "workload" || strings.HasPrefix(cells[0], "---") {
			continue
		}
		allocs, err1 := strconv.ParseUint(cells[1], 10, 64)
		bytes, err2 := strconv.ParseUint(cells[2], 10, 64)
		if err1 != nil || err2 != nil {
			t.Fatalf("%s: bad numbers in row %q", baselineFile, cells[0])
		}
		out[cells[0]] = metric{allocs, bytes}
	}
	return out
}

func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	parts := strings.Split(strings.Trim(line, "|"), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func writeBaseline(t *testing.T, path string, names []string, got map[string]metric) {
	t.Helper()
	var table strings.Builder
	table.WriteString("| workload | allocs | bytes |\n|---|---:|---:|\n")
	for _, name := range names {
		fmt.Fprintf(&table, "| %s | %d | %d |\n", name, got[name].allocs, got[name].bytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	begin := strings.Index(text, beginMarker)
	end := strings.Index(text, endMarker)
	if begin < 0 || end < 0 {
		t.Fatalf("%s: missing %s block", baselineFile, beginMarker)
	}
	out := text[:begin] + beginMarker + "\n" + table.String() + text[end:]
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
}
