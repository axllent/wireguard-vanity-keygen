package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

// Checks the search does not contain any invalid characters
func isValidSearch(s string) bool {
	var r = regexp.MustCompile(`[^a-zA-Z0-9\/\+]`)
	return !r.MatchString(s)
}

// Returns a human-readable output of time.Duration
func humanizeDuration(duration time.Duration) string {
	// more than duration can handle
	if duration.Hours() < 0.0 {
		return fmt.Sprintf("hundreds of years")
	}
	if duration.Hours() > 8760.0 {
		y := int64(duration.Hours() / 8760)
		return fmt.Sprintf("%d %s", y, plural("year", y))
	}
	if duration.Hours() > 720.0 {
		m := int64(duration.Hours() / 24 / 30)
		return fmt.Sprintf("%d %s", m, plural("month", m))
	}
	if duration.Hours() > 168.0 {
		w := int64(duration.Hours() / 168)
		return fmt.Sprintf("%d %s", w, plural("week", w))
	}
	if duration.Seconds() < 60.0 {
		s := int64(duration.Seconds())
		return fmt.Sprintf("%d %s", s, plural("second", s))
	}
	if duration.Minutes() < 60.0 {
		m := int64(duration.Minutes())
		return fmt.Sprintf("%d %s", m, plural("minute", m))
	}
	if duration.Hours() < 24.0 {
		m := int64(math.Mod(duration.Minutes(), 60))
		h := int64(duration.Hours())
		return fmt.Sprintf("%d %s, %d %s",
			h, plural("hour", h), m, plural("minute", m))
	}
	// if duration.Hours() <
	h := int64(math.Mod(duration.Hours(), 24))
	d := int64(duration.Hours() / 24)
	return fmt.Sprintf("%d %s, %d %s",
		d, plural("day", d), h, plural("hour", h))
}

// Returns a plural of `s` if the value `v` is 0 or > 0
func plural(s string, v int64) string {
	if v == 1 {
		return s
	}

	return s + "s"
}

// Returns a number-formatted string, eg: 1,123,456
func numberFormat(n int64) string {
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
