package keygen

import (
	"testing"
)

// BenchmarkKeygenGenerationSpeed benchmarks the speed of generating new WireGuard private keys.
func BenchmarkKeygenGenerationSpeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := newPrivateKey()
		if err != nil {
			b.Fatalf("failed to generate private key: %v", err)
		}
	}
}
