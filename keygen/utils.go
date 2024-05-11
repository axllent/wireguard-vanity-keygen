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

// regexWillNeverMatch is a shared error message that the regex will never match
const regexWillNeverMatch = "The regular expression will never match"

// IsValidSearch checks the search does not contain any invalid characters
func IsValidSearch(s string) bool {
	var r = regexp.MustCompile(`[^a-zA-Z0-9\/\+]`)
	return !r.MatchString(s)
}

// InvalidSearchMsg returns the error message the search term contains invalid characters
func InvalidSearchMsg(s string) string {
	return fmt.Sprintf("\n\"%s\" contains invalid characters\nValid characters include letters [a-z], numbers [0-9], + and /", s)
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

// IsRegex returns true if any regex metacharacters (except +) are in the search term
func IsRegex(s string) bool {
	return strings.ContainsAny(s, regexChars)
}

// invalidRegexMsg returns an error message how the regex is invalid
func invalidRegexMsg(s string, errmsg string) string {
	return fmt.Sprintf("\n\"%s\" is an invalid regular expression\n%s", s, errmsg)
}

// IsValidRegex checks the regex has any chance of matching a key
func IsValidRegex(s string) string {
	// A consise guide on golang's regex syntax is at
	// https://pkg.go.dev/regexp/syntax

	stripped := removeMetacharacters(s)
	if !IsValidSearch(stripped) {
		return InvalidSearchMsg(s)
	}

	// Expressions with '^' character
	re := regexp.MustCompile(`.\^`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, "The '^' character must appear at the beginning of the search term")
	}

	// Expressions with '$' character
	re = regexp.MustCompile(`\$.`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, "The '$' character must appear at the end of the search term")
	}
	re = regexp.MustCompile(`[^=]\$`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, "A search at the end of the string must contain an '=' character, as all keys end with an `=`")
	}
	re = regexp.MustCompile(`=[^$]`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, "The '=' character can only appear at the end of a key")
	}
	// The command:
	// wireguard-vanity-keygen -l 1000 . | grep private | cut -c 105- | sort -u | tr -d "=" | tr -d "\n"
	// outputs:
	// 048AEIMQUYcgkosw
	re = regexp.MustCompile(`[^048AEIMQUYcgkosw]=\$`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	// Expressions with backslashes:

	// A regex of just a backslash and a single character will never match
	re = regexp.MustCompile(`^\\.$`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	// Control characters and many octal values will meter match, disallow them all
	re = regexp.MustCompile(`\\[aftnrxswWpP0-7]`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	// Disallow backslashes followed by any non-alnum or + character
	re = regexp.MustCompile(`\\[^A-Za-z0-9+]`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	// Expressions with character classes: [[:alnum:]], etc.

	// [[:blank:]], [[:cntrl:]], [[:punct:]] and [[:space:]] will never match
	re = regexp.MustCompile(`\[\[:(blank|cntrl|punct|space):\]\]`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	// [[^:ascii:]], [[^:graph:]], [[^:print:]] will never match
	re = regexp.MustCompile(`\[\[\^:(ascii|graph|print):\]\]`)
	if re.MatchString(s) {
		return invalidRegexMsg(s, regexWillNeverMatch)
	}

	return ""
}

// removeMetacharacters removes regex metacharacters from the string
func removeMetacharacters(s string) string {
	// This logic isn't needed anymore, as we don't attempt to calculate the probability of regular expressions
	// // remove (?i) from beginning of string
	// re := regexp.MustCompile(`^\([^)]*\)`)
	// s = re.ReplaceAllLiteralString(s, "")
	// // replace [a-b]+ with x
	// re = regexp.MustCompile(`\[[^]]*\]\+?`)
	// s = re.ReplaceAllLiteralString(s, "x")
	// // strip all {n}
	// re = regexp.MustCompile(`\{[^}]+\}`)
	// s = re.ReplaceAllLiteralString(s, "")
	// // replace = with x
	// re = regexp.MustCompile(`=`)
	// s = re.ReplaceAllLiteralString(s, "x")

	// strip out remaining regexp metacharacters
	for _, rune1 := range []rune(regexChars) {
		s = strings.ReplaceAll(s, string(rune1), "")
	}
	return s
}
