// Package main is the main package for the WireGuard Vanity Key Generator application.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/axllent/ghru/v2"
	"github.com/axllent/wireguard-vanity-keygen/keygen"
	"github.com/spf13/pflag"
)

var (
	options    keygen.Options
	appVersion = "dev"
	ghruConf   = ghru.Config{
		Repo:           "axllent/wireguard-vanity-keygen",
		ArchiveName:    "wireguard-vanity-keygen-{{.OS}}-{{.Arch}}",
		BinaryName:     "wireguard-vanity-keygen",
		CurrentVersion: appVersion,
	}
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

	var summary, showVersion, update bool
	var jsonFile string
	flag.BoolVarP(&summary, "summary", "s", false, "print results when all are found (default false)")
	flag.BoolVarP(&options.CaseSensitive, "case-sensitive", "c", false, "case sensitive match (default false)")
	flag.IntVarP(&options.Threads, "threads", "t", options.Cores, "threads")
	flag.IntVarP(&options.LimitResults, "limit", "l", 1, "limit results to n (exists after)")
	flag.StringVarP(&options.Timeout, "timeout", "T", "", "quit after n minutes (allowed suffixes: s/m/h) (default \"\")")
	flag.StringVarP(&jsonFile, "json", "j", "", "write results to JSON file")
	flag.BoolVarP(&showVersion, "version", "v", false, "show app version")
	flag.BoolVarP(&update, "update", "u", false, "update to latest release")

	flag.Parse(os.Args[1:])
	args := flag.Args()

	if showVersion {
		fmt.Printf("Version: %s\n", appVersion)
		release, err := ghruConf.Latest()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if release.Tag != appVersion {
			fmt.Printf(
				"Update available: %s\nRun `%s -u` to update (requires read/write access to install directory).\n",
				release.Tag,
				os.Args[0],
			)
		}
		os.Exit(0)
	}

	if update {
		rel, err := ghruConf.SelfUpdate()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel.Tag)
		os.Exit(0)
	}

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
			c.WordMap[sword] = &keygen.AtomicCounter{Value: int64(options.LimitResults)}
			probability := keygen.CalculateProbability(sword, options.CaseSensitive)
			estimate64 := int64(speed) * probability
			estimate := time.Duration(estimate64)

			fmt.Printf("Probability for \"%s\": 1 in %s (approx %s per match)\n",
				word, keygen.NumberFormat(probability), keygen.HumanizeDuration(estimate))
			continue
		}

		errMsg := keygen.IsValidRegex(sword)
		if errMsg != "" {
			fmt.Fprintln(os.Stderr, errMsg)
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
		c.RegexpMap[re] = &keygen.AtomicCounter{Value: int64(options.LimitResults)}
	}

	if timeout > time.Duration(0) {
		fmt.Printf("\nQuitting after %v, or sooner if all matching keys are found...\n", timeout)
	}

	fmt.Printf("\nPress Ctrl-c to cancel\n\n")

	var results []keygen.Pair
	if !summary && jsonFile == "" {
		c.Find(func(match keygen.Pair) {
			fmt.Printf("private: %s   public: %s\n", match.Private, match.Public)
		})
	} else {
		results = c.CollectToSlice()
		for _, match := range results {
			fmt.Printf("private: %s   public: %s\n", match.Private, match.Public)
		}
	}

	if jsonFile != "" {
		jsonFile = filepath.Clean(jsonFile)
		if results == nil {
			results = []keygen.Pair{}
		}
		data, err := json.MarshalIndent(struct {
			Results []keygen.Pair `json:"results"`
		}{Results: results}, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(jsonFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nResults written to %s\n", jsonFile)
	}
}

// parseTimeout parses the timeout string to a time.Duration. If the input is
// solely digits, minutes is assumed
func parseTimeout(t string) (time.Duration, error) {
	if t == "" {
		return time.Duration(0), nil
	}

	re := regexp.MustCompile(`^[\d\.]+$`)
	if re.MatchString(t) {
		t += "m"
	}
	duration, err := time.ParseDuration(t)
	if err != nil {
		return time.Duration(0), err
	}

	return duration, nil
}
