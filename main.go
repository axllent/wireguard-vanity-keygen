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
	flag.StringVarP(&options.Timeout, "timeout", "T", "", "quit after n minutes (allowed suffixes: s/m/h) (default \"\")")

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

	timeout, err := parseTimeout(options.Timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid timeout value: %s\n", err)
		os.Exit(2)
	}

	if timeout > time.Duration(0) {
		fmt.Printf("Quitting after %v\n", timeout)
	}

	c := keygen.New(options, timeout)

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
		word = strings.Trim(word, " ")
		sword := word
		if !keygen.IsRegex(sword) {
			if !keygen.IsValidSearch(sword) {
				fmt.Fprintln(os.Stderr, keygen.InvalidSearchMsg(word))
				os.Exit(2)
			}
			if !options.CaseSensitive {
				sword = strings.ToLower(sword)
			}
			c.WordMap[sword] = options.LimitResults
			probability := keygen.CalculateProbability(sword, options.CaseSensitive)
			estimate64 := int64(speed) * probability
			estimate := time.Duration(estimate64)

			fmt.Printf("Probability for \"%s\": 1 in %s (approx %s per match)\n",
				word, keygen.NumberFormat(probability), keygen.HumanizeDuration(estimate))
			continue
		}

		errmsg := keygen.IsValidRegex(sword)
		if errmsg != "" {
			fmt.Fprintln(os.Stderr, errmsg)
			os.Exit(2)
		}

		fmt.Printf("Probability for \"%s\" cannot be calculated as it is a regular expression\n", sword)

		// strip off leading .* as it's implied:
		re := regexp.MustCompile(`^\.\*`)
		sword = re.ReplaceAllLiteralString(sword, "")
		// strip off trailing .* as it's implied:
		re = regexp.MustCompile(`\.\*$`)
		sword = re.ReplaceAllLiteralString(sword, "")

		regex := sword
		if !options.CaseSensitive {
			if !strings.HasPrefix(regex, "(?i)") {
				regex = "(?i)" + regex
			}
		}
		re, err := regexp.Compile(regex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n\"%s\" is an invalid regular expression: %v\n", word, err)
			os.Exit(2)
		}
		c.RegexpMap[re] = options.LimitResults
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

// parseTimeout parses the timeout string to a time.Duration. If the input is
// solely digits, minutes is assumed
func parseTimeout(t string) (time.Duration, error) {
	if t == "" {
		return time.Duration(0), nil
	}

	re := regexp.MustCompile(`^\d+$`)
	if re.MatchString(t) {
		t += "m"
	}
	complex, err := time.ParseDuration(t)
	if err != nil {
		return time.Duration(0), err
	}

	return complex, nil
}
