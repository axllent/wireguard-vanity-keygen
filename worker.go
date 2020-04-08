package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	wordMap  map[string]int
	mapMutex = sync.RWMutex{}
)

// Crunch will generate a new key and compare to the search(s)
func Crunch() {
	k, err := newPrivateKey()
	if err != nil {
		panic(err)
	}

	pub := k.Public().String()
	matchKey := pub

	if !caseSensitive {
		matchKey = strings.ToLower(pub)
	}

	// Assume the task is completed, once all searched have been found
	// and limits have been reached, it sends an exit signal
	completed := true

	// Allow only one routine at a time to avoid
	// "concurrent map iteration and map write"
	mapMutex.Lock()
	defer mapMutex.Unlock()
	for w, count := range wordMap {
		if count == 0 {
			continue
		}
		completed = false
		if strings.HasPrefix(matchKey, w) {
			wordMap[w] = count - 1
			fmt.Printf("private %s   public %s\n", k.String(), pub)
		}
	}

	if completed {
		// send exit status, allows time for processes to exit
		stopChan <- 1
	}

	<-threadChan // removes an int from threads, allowing another to proceed
}

// CalculateSpeed returns average calculations per second based
// on the time per run taken from 2 seconds runtime.
func calculateSpeed() (int64, time.Duration) {
	var n int64
	n = 1
	start := time.Now()

	for loopStart := time.Now(); ; {
		// check every 100 loops if time is reached
		if n%100 == 0 {
			if time.Since(loopStart) > 2*time.Second {
				break
			}
		}

		threadChan <- 1 // will block if there is MAX ints in threads
		go func() {
			// dry run
			k, err := newPrivateKey()
			if err != nil {
				panic(err)
			}
			_ = k.String()
			t := strings.ToLower(k.Public().String())

			// Allow only one routine at a time to avoid
			// "concurrent map iteration and map write"
			mapMutex.Lock()
			defer mapMutex.Unlock()
			for w := range wordMap {
				_ = strings.HasPrefix(t, w)
			}
			<-threadChan // removes an int from threads, allowing another to proceed
			n++
		}()
	}

	estimate64 := int64(time.Since(start)) / n

	return n / 2, time.Duration(estimate64)
}

// CalculateProbability calculates the probability that a string
// can be found. Case-insensitive letter matches [a-z] can be
// found in upper and lowercase combinations, so have a higher
// chance of being found than [0-9], / or +, or case-sensitive matches.
func calculateProbability(s string) int64 {
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
