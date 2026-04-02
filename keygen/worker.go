package keygen

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Options struct
type Options struct {
	LimitResults  int
	Threads       int
	CaseSensitive bool
	Cores         int
	Timeout       string
}

// Cruncher struct

// AtomicCounter struct
type AtomicCounter struct {
	Value int64
}

// Dec decrements the counter and returns the new value
func (a *AtomicCounter) Dec() int64 {
	return atomic.AddInt64(&a.Value, -1)
}

// Get returns the current value of the counter
func (a *AtomicCounter) Get() int64 {
	return atomic.LoadInt64(&a.Value)
}

// Cruncher struct
type Cruncher struct {
	Options
	WordMap   map[string]*AtomicCounter
	RegexpMap map[*regexp.Regexp]*AtomicCounter
	Abort     bool // set to true to abort processing
	timeout   time.Duration
	timedOut  bool
}

// Pair struct
type Pair struct {
	Private string
	Public  string
}

// New returns a Cruncher
func New(options Options, timeout time.Duration) *Cruncher {
	return &Cruncher{
		Options:   options,
		WordMap:   make(map[string]*AtomicCounter),
		RegexpMap: make(map[*regexp.Regexp]*AtomicCounter),
		timeout:   timeout,
	}
}

// Crunch will generate a new key and compare to the search(s)
func (c *Cruncher) crunch(cb func(match Pair)) bool {
	k, err := newPrivateKey()
	if err != nil {
		panic(err)
	}

	pub := k.Public().String()
	matchKey := pub

	if !c.CaseSensitive {
		matchKey = strings.ToLower(pub)
	}

	completed := true

	for w, counter := range c.WordMap {
		if counter.Get() <= 0 {
			continue
		}
		completed = false
		if strings.HasPrefix(matchKey, w) {
			if counter.Dec() >= 0 {
				cb(Pair{Private: k.String(), Public: pub})
			}
		}
	}

	for w, counter := range c.RegexpMap {
		if counter.Get() <= 0 {
			continue
		}
		completed = false
		if w.MatchString(matchKey) {
			if counter.Dec() >= 0 {
				cb(Pair{Private: k.String(), Public: pub})
			}
		}
	}

	return completed
}

// CalculateSpeed returns average calculations per second based
// on the time per run taken from 2 seconds runtime.
func (c *Cruncher) CalculateSpeed() (int64, time.Duration) {
	var n int64
	atomic.StoreInt64(&n, 1)
	start := time.Now()
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < c.Cores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				k, err := newPrivateKey()
				if err != nil {
					panic(err)
				}
				_ = k.String()
				t := strings.ToLower(k.Public().String())
				for w := range c.WordMap {
					_ = strings.HasPrefix(t, w)
				}
				for w := range c.RegexpMap {
					_ = w.MatchString(t)
				}
				atomic.AddInt64(&n, 1)
			}
		}()
	}

	time.Sleep(2 * time.Second)
	close(done)
	wg.Wait()

	total := atomic.LoadInt64(&n)
	elapsed := time.Since(start)
	estimate := time.Duration(int64(elapsed) / total)
	return total / 2, estimate
}

// CalculateProbability calculates the probability that a string
// can be found. Case-insensitive letter matches [a-z] can be
// found in upper and lowercase combinations, so have a higher
// chance of being found than [0-9], / or +, or case-sensitive matches.
func CalculateProbability(s string, caseSensitive bool) int64 {
	var nonAlphaProbability, alphaProbability int64
	alphaProbability = 26 + 10 + 2
	nonAlphaProbability = 26 + 26 + 10 + 2
	if caseSensitive {
		alphaProbability = nonAlphaProbability
	}
	ascii := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var p int64
	p = 1

	for _, char := range s {
		if !strings.Contains(ascii, string(char)) {
			p = p * nonAlphaProbability
		} else {
			p = p * alphaProbability
		}
	}

	return p
}

// CollectToSlice will run till all the matching keys were calculated. This can take some time
func (c *Cruncher) CollectToSlice() []Pair {
	var matches []Pair
	c.Find(func(match Pair) {
		matches = append(matches, match)
	})
	return matches
}

// Find will invoke a callback function for each match to support some interactivity or at least feedback
func (c *Cruncher) Find(cb func(match Pair)) {
	var wg sync.WaitGroup

	if c.timeout == time.Duration(0) {
		for i := 0; i < c.Cores; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for !c.Abort {
					if c.crunch(cb) {
						c.Abort = true
						return
					}
				}
			}()
		}
		wg.Wait()
		return
	}

	t := time.NewTimer(c.timeout)
	defer t.Stop()

	for i := 0; i < c.Cores; i++ {
		wg.Add(1)
		go func(t *time.Timer) {
			defer wg.Done()
			for !c.Abort {
				if c.crunch(cb) {
					c.Abort = true
					return
				}
				select {
				case <-t.C:
					c.timedOut = true
					c.Abort = true
					return
				default:
				}
			}
		}(t)
	}
	wg.Wait()

	if c.timedOut {
		fmt.Printf("Timed out after %v\n", c.timeout)
	}
}
