package keygen

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// regexChars contains the list of regex metacharacters, excluding +,
// which is valid in a key
const regexChars = `^$.|?*-[]{}()\`

// IsValidSearch checks the search does not contain any invalid characters
func IsValidSearch(s string) bool {
	var r = regexp.MustCompile(`[^a-zA-Z0-9\/\+]`)
	return !r.MatchString(s)
}

// HumanizeDuration returns a human-readable output of time.Duration
func HumanizeDuration(duration time.Duration) string {
	// more than duration can handle
	if duration.Hours() < 0.0 {
		return fmt.Sprintf("hundreds of years")
	}
	if duration.Hours() > 8760.0 {
		y := int64(duration.Hours() / 8760)
		return fmt.Sprintf("%d %s", y, Plural("year", y))
	}
	if duration.Hours() > 720.0 {
		m := int64(duration.Hours() / 24 / 30)
		return fmt.Sprintf("%d %s", m, Plural("month", m))
	}
	if duration.Hours() > 168.0 {
		w := int64(duration.Hours() / 168)
		return fmt.Sprintf("%d %s", w, Plural("week", w))
	}
	if duration.Seconds() < 60.0 {
		s := int64(duration.Seconds())
		return fmt.Sprintf("%d %s", s, Plural("second", s))
	}
	if duration.Minutes() < 60.0 {
		m := int64(duration.Minutes())
		return fmt.Sprintf("%d %s", m, Plural("minute", m))
	}
	if duration.Hours() < 24.0 {
		m := int64(math.Mod(duration.Minutes(), 60))
		h := int64(duration.Hours())
		return fmt.Sprintf("%d %s, %d %s",
			h, Plural("hour", h), m, Plural("minute", m))
	}
	// if duration.Hours() <
	h := int64(math.Mod(duration.Hours(), 24))
	d := int64(duration.Hours() / 24)
	return fmt.Sprintf("%d %s, %d %s",
		d, Plural("day", d), h, Plural("hour", h))
}

// Plural returns a Plural of `s` if the value `v` is 0 or > 0
func Plural(s string, v int64) string {
	if v == 1 {
		return s
	}

	return s + "s"
}

// NumberFormat returns a number-formatted string, eg: 1,123,456
func NumberFormat(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func RemoveMetacharacters(s string) string {
	if !strings.ContainsAny(s, regexChars) {
		return s
	}
	// remove (?i)
	re1 := regexp.MustCompile(`^\([^)]*\)`)
	s = re1.ReplaceAllLiteralString(s, "")
	// replace [a-b]+ with a
	re2 := regexp.MustCompile(`\[[^]]*\]\+?`)
	s = re2.ReplaceAllLiteralString(s, "a")
	for _, rune1 := range []rune(regexChars) {
		s = strings.ReplaceAll(s, string(rune1), "")
	}
	return s
}
