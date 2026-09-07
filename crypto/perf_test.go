package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
)

// Message sizes spanning the shipped traffic profile: a small control
// message, a typical application payload, and a datagram-sized chunk.
var benchSizes = []int{64, 1024, 16384}

func benchChannel(tb testing.TB) *Channel {
	sendKey := make([]byte, 32)
	recvKey := make([]byte, 32)
	for i := range sendKey {
		sendKey[i] = byte(i)
		recvKey[i] = byte(255 - i)
	}
	ch, err := NewChannel(sendKey, recvKey)
	if err != nil {
		tb.Fatal(err)
	}
	return ch
}

func benchPayload(n int) []byte {
	p := make([]byte, n)
	if _, err := rand.Read(p); err != nil {
		panic(err)
	}
	return p
}

// Every relayed byte goes through Encrypt, so its per-message allocation
// behaviour is the one that matters most in this repository.
func BenchmarkChannelEncrypt(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			ch := benchChannel(b)
			payload := benchPayload(n)
			b.SetBytes(int64(n))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ch.Encrypt(payload)
			}
		})
	}
}

func BenchmarkChannelDecrypt(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			// Sender and receiver must be counterparts: the sender's send
			// key is the receiver's recv key.
			enc := benchChannel(b)
			dec, err := NewChannel(
				bytesOf(func(i int) byte { return byte(255 - i) }),
				bytesOf(func(i int) byte { return byte(i) }))
			if err != nil {
				b.Fatal(err)
			}
			_ = dec
			payload := benchPayload(n)
			msgs := make([][]byte, b.N)
			for i := range msgs {
				msgs[i] = enc.Encrypt(payload)
			}
			recv, err := NewChannel(
				bytesOf(func(i int) byte { return byte(255 - i) }),
				bytesOf(func(i int) byte { return byte(i) }))
			if err != nil {
				b.Fatal(err)
			}
			b.SetBytes(int64(n))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := recv.Decrypt(msgs[i]); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func bytesOf(f func(int) byte) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = f(i)
	}
	return b
}

// Golden ciphertexts captured from the implementation before the
// buffer-management change, on a fixed key and a fixed starting sequence
// number. Encrypt feeds Swift, Kotlin, C and browser peers, so its output
// has to stay byte-identical, not merely decryptable by itself.
var encryptGolden = []struct {
	size int
	hex  string
}{
	{0, "00000000000000004af60e46eeff381ceef6c0246c0dde69"},
	{1, "0100000000000000e27afcf49bdad9bf3c65af3c4271bf4677"},
	{64, "0200000000000000e669e575e686428c95822859d8dd956b217b16455a681ab2e2d9ae87ac4aabcaf513a533974d115ab84d4fd24935e9c7793d1a8a24fbe652268e7845eff21a8c7e87f15d8d5807fced109f598d0b7b66"},
}

func TestEncryptMatchesGoldenBytes(t *testing.T) {
	ch := benchChannel(t)
	for i, want := range encryptGolden {
		p := make([]byte, want.size)
		for j := range p {
			p[j] = byte(j * 7)
		}
		got := hex.EncodeToString(ch.Encrypt(p))
		if got != want.hex {
			t.Fatalf("message %d (%d bytes):\n got %s\nwant %s", i, want.size, got, want.hex)
		}
	}
}

// Encrypt and Decrypt must remain exact counterparts across the wire and
// across languages, so a buffer-management change has to preserve bytes
// exactly, not merely round-trip.
func TestEncryptIsDeterministicForAGivenSequence(t *testing.T) {
	payload := []byte("the quick brown fox jumps over the lazy dog")
	first := benchChannel(t).Encrypt(payload)
	second := benchChannel(t).Encrypt(payload)
	if !bytes.Equal(first, second) {
		t.Fatalf("same key and sequence produced different ciphertext:\n%x\n%x", first, second)
	}
	if len(first) != 8+len(payload)+16 {
		t.Fatalf("unexpected ciphertext length %d for %d-byte payload", len(first), len(payload))
	}
}
