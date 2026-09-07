# Performance baseline

## What is gated, and what is only reported

`make bench` runs the Go benchmarks in `crypto/perf_test.go` and prints
throughput. Those numbers are useful for judging a change and useless as
a gate: on a machine with anything else running they move by tens of
percent between runs of the same binary.

`TestRatchetChannelAllocations` in `crypto/` locks allocation counts and
bytes instead. Those do not vary with machine load. It runs inside the
ordinary `go test ./...`, so `make bullseye` enforces it with no extra
step.

The band is ±1 % in **both** directions. A regression fails, and an
improvement also fails until it is re-recorded in the same commit:

```
PIGEON_PERF_RECORD=1 go test ./crypto/ -run Ratchet
```

A workload with no baseline row fails rather than running unmeasured.

<!-- perf-baseline:begin -->
| workload | allocs | bytes |
|---|---:|---:|
| decrypt/1024 | 2 | 1040 |
| decrypt/16384 | 2 | 16400 |
| decrypt/64 | 2 | 80 |
| encrypt/1024 | 3 | 1184 |
| encrypt/16384 | 3 | 18464 |
| encrypt/64 | 3 | 128 |
<!-- perf-baseline:end -->

## Why the channel, and not something else

Every byte the relay carries is sealed by `crypto.Channel.Encrypt` and
opened by `Decrypt`. The relay is explicitly not trusted to see plaintext,
so there is no path that skips them. If any allocation behaviour in this
repository is worth locking, it is theirs.

## Encrypt: 2.25 bytes of heap per byte relayed

`Encrypt` used to seal into a fresh buffer and then append that buffer to
the 8-byte sequence prefix:

```go
seqBytes := make([]byte, 8)
ciphertext := c.sendGCM.Seal(nil, nonce, plaintext, seqBytes)
return append(seqBytes, ciphertext...)
```

That allocates the ciphertext once inside `Seal`, allocates it again when
`append` outgrows the 8-byte prefix, and copies it between the two. It now
allocates one buffer at its exact final size and seals directly into it.

Measured with `go test -bench ChannelEncrypt -benchmem -benchtime=200x`:

| Payload | Before | After |
|---|---|---|
| 64 B | 200 B/op, 4 allocs | 120 B/op, 3 allocs |
| 1 KiB | 2,328 B/op, 4 allocs | 1,176 B/op, 3 allocs |
| 16 KiB | 36,888 B/op, 4 allocs | 18,456 B/op, 3 allocs |

Throughput on the 16 KiB case went from 2,826 MB/s to 4,471 MB/s in the
same runs, but that figure was taken under load and is not the claim; the
byte counts are.

The remaining third allocation is the nonce, from `makeNonce`, which
escapes into the AEAD interface. `Decrypt` was already minimal at two
allocations and one output buffer, and is unchanged.

## Output compatibility

`Encrypt` feeds Swift, Kotlin, C and browser peers, so a buffer change
has to preserve bytes exactly rather than merely round-trip.
`TestEncryptMatchesGoldenBytes` pins three ciphertexts, captured from the
implementation before the change, on a fixed key and fixed starting
sequence number. It passed before the change and passes after it.

## Reference host

Apple M4 Max, 16 cores, macOS 26, Go 1.26. Allocation counts are
host-independent in principle; byte totals depend on the runtime's size
classes, so they are recorded on one host.
