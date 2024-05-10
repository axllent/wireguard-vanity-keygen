package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/axllent/wireguard-vanity-keygen/keygen"
	"github.com/spf13/pflag"
)

var (
	options    keygen.Options
	appVersion = "dev"
)

func main() {

	flag := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	// detect number of cores minus one
	options.Cores = runtime.NumCPU() - 1
	if options.Cores == 0 {
		// if it is single-code then it
		options.Cores = 1
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

	var summary bool
	flag.BoolVarP(&summary, "summary", "s", false, "print results when all are found (default false)")
	flag.BoolVarP(&options.CaseSensitive, "case-sensitive", "c", false, "case sensitive match (default false)")
	flag.IntVarP(&options.Threads, "threads", "t", options.Cores, "threads")
	flag.IntVarP(&options.LimitResults, "limit", "l", 1, "limit results to n (exists after)")

	flag.Parse(os.Args[1:])
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
	}

	if options.Threads == 0 || options.Threads > options.Cores {
		fmt.Printf("invalid number of cores: %d\n", options.Threads)
		flag.Usage()
	} else {
		options.Cores = options.Threads
	}

	c := keygen.New(options)

	fmt.Printf("Calculating speed: ")

	perSecond, speed := c.CalculateSpeed()
	fmt.Printf("%s calculations per second using %d CPU %s\n", keygen.NumberFormat(perSecond), options.Cores, keygen.Plural("core", int64(options.Cores)))

	cs := "insensitive"
	if options.CaseSensitive {
		cs = "sensitive"
	}
	fmt.Printf("Case-%s search, exiting after %d %s\n",
		cs, options.LimitResults, keygen.Plural("result", int64(options.LimitResults)))

	for _, word := range args {
		sword := word
		if !options.CaseSensitive {
			sword = strings.ToLower(sword)
		}
		stripped := keygen.RemoveMetacharacters(sword)
		if !keygen.IsValidSearch(stripped) {
			fmt.Printf("\n\"%s\" contains invalid characters\n", word)
			fmt.Println("Valid characters include letters [a-z], numbers [0-9], + and /")
			os.Exit(2)
		}
		if stripped != sword {
			regex := regexp.MustCompile(sword)
			c.RegexpMap[regex] = options.LimitResults
		} else {
			c.WordMap[sword] = options.LimitResults
		}
		probability := keygen.CalculateProbability(stripped, options.CaseSensitive)
		estimate64 := int64(speed) * probability
		estimate := time.Duration(estimate64)

		comment := ""
		if len(stripped) != len(sword) {
			comment = fmt.Sprintf(" (approximation may be wildly off, as '%s' is test string)", stripped)
		}
		fmt.Printf("Probability for \"%s\": 1 in %s (approx %s per match)%s\n",
			word, keygen.NumberFormat(probability), keygen.HumanizeDuration(estimate), comment)
	}

	fmt.Printf("\nPress Ctrl-c to cancel\n\n")
	if !summary {
		c.Find(func(match keygen.Pair) {
			fmt.Printf("private %s   public %s\n", match.Private, match.Public)
		})
	} else {
		for _, match := range c.CollectToSlice() {
			fmt.Printf("private %s   public %s\n", match.Private, match.Public)
		}
	}
}
