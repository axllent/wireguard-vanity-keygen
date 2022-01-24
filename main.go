package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	limitResults  int
	caseSensitive bool
	threadChan    chan int
	stopChan      chan int
	appVersion    = "dev"
)

func main() {
	var userCores int

	flag := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	// detect number of cores minus one
	cores := runtime.NumCPU() - 1
	if cores == 0 {
		// if it is single-code then it
		cores = 1
	}

	// set the default help
	flag.Usage = func() {
		fmt.Printf("WireGuard Vanity Key Generator (%s)\n\n", appVersion)
		// fmt.Printf("Version: %s\n\n", appVersion)
		fmt.Printf("Usage: %s [OPTIONS] <SEARCH> [<SEARCH>...]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.SortFlags = false
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("https://github.com/axllent/wireguard-vanity-keygen")
		fmt.Println()
		os.Exit(0)
	}

	flag.BoolVarP(&caseSensitive, "case-sensitive", "c", false, "case sensitive match (default false)")
	flag.IntVarP(&userCores, "threads", "t", cores, "threads")
	flag.IntVarP(&limitResults, "limit", "l", 1, "limit results to n (exists after)")

	flag.Parse(os.Args[1:])
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
	}

	if userCores == 0 || userCores > cores {
		fmt.Printf("invalid number of cores: %d\n", userCores)
		flag.Usage()
	} else {
		cores = userCores
	}

	// CPU thread channel
	threadChan = make(chan int, cores)

	// construct the wordmap
	wordMap = make(map[string]int)

	fmt.Printf("Calculating speed: ")

	perSecond, speed := calculateSpeed()
	fmt.Printf("%s calculations per second using %d CPU %s\n", numberFormat(perSecond), cores, plural("core", int64(cores)))

	cs := "insensitive"
	if caseSensitive {
		cs = "sensitive"
	}
	fmt.Printf("Case-%s search, exiting after %d %s\n",
		cs, limitResults, plural("result", int64(limitResults)))

	for _, word := range args {
		sword := word
		if !caseSensitive {
			sword = strings.ToLower(sword)
		}
		if !isValidSearch(sword) {
			fmt.Printf("\n\"%s\" contains invalid characaters\n", word)
			fmt.Println("Valid characters include letters [a-z], numbers [0-9], + and /")
			os.Exit(2)
		}
		wordMap[sword] = limitResults

		probability := calculateProbability(sword)
		estimate64 := int64(speed) * probability
		estimate := time.Duration(estimate64)

		fmt.Printf("Probability for \"%s\": 1 in %s (approx %s per match)\n",
			word, numberFormat(probability), humanizeDuration(estimate))
	}

	stopChan = make(chan int)

	go func() {
		_ = <-stopChan
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	fmt.Printf("\nPress Ctrl-c to cancel\n\n")

	for {
		threadChan <- 1 // will block if there is MAX ints in threads
		go Crunch()
	}
}
