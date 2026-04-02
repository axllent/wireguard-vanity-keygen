package keygen

import (
	"encoding/base64"
	"math"
	"runtime"
	"sync"
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

// BenchmarkCrunchThroughput benchmarks concurrent crunch() throughput including
// key generation, base64 encoding, case conversion, and prefix matching.
// This reflects the worker pool design: fixed goroutines loop internally.
func BenchmarkCrunchThroughput(b *testing.B) {
	opts := Options{Cores: runtime.NumCPU(), CaseSensitive: false}
	c := New(opts, 0)
	// Use a prefix that will never match so the counter never saturates
	c.WordMap["zzzz"] = &AtomicCounter{Value: math.MaxInt64}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		buf := make([]byte, base64.StdEncoding.EncodedLen(KeySize)) // one per goroutine
		for pb.Next() {
			c.crunch(func(Pair) {}, buf)
		}
	})
}

// BenchmarkGoroutinePerAttempt simulates the previous design where a new goroutine
// was spawned for every single key attempt. Compare against BenchmarkCrunchThroughput
// to quantify the goroutine lifecycle overhead that the worker pool eliminates.
func BenchmarkGoroutinePerAttempt(b *testing.B) {
	cores := runtime.NumCPU()
	thread := make(chan int, cores)
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		thread <- 1
		wg.Add(1)
		go func() {
			defer wg.Done()
			k, err := newPrivateKey()
			if err != nil {
				panic(err)
			}
			_ = k.Public().String()
			<-thread
		}()
	}
	wg.Wait()
}
