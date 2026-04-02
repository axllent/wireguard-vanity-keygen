package keygen

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"
	"time"
)

// --- crypto.go ---

func TestNewPrivateKey(t *testing.T) {
	k, err := newPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Clamping per RFC 7748
	if k[0]&7 != 0 {
		t.Errorf("low 3 bits of k[0] not cleared: %08b", k[0])
	}
	if k[31]&128 != 0 {
		t.Errorf("high bit of k[31] not cleared: %08b", k[31])
	}
	if k[31]&64 == 0 {
		t.Errorf("bit 6 of k[31] not set: %08b", k[31])
	}
}

func TestPrivateKeyString(t *testing.T) {
	k, err := newPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := k.String()
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("private key string is not valid base64: %v", err)
	}
	if len(decoded) != KeySize {
		t.Errorf("expected %d bytes, got %d", KeySize, len(decoded))
	}
}

func TestPublicKeyString(t *testing.T) {
	k, err := newPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pub := k.Public()
	s := pub.String()
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("public key string is not valid base64: %v", err)
	}
	if len(decoded) != KeySize {
		t.Errorf("expected %d bytes, got %d", KeySize, len(decoded))
	}
}

func TestPublicKeyDeterministic(t *testing.T) {
	k, err := newPrivateKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pub1 := k.Public()
	pub2 := k.Public()
	if pub1 != pub2 {
		t.Error("Public() is not deterministic for the same private key")
	}
}

func TestDistinctKeysGenerated(t *testing.T) {
	k1, _ := newPrivateKey()
	k2, _ := newPrivateKey()
	if k1 == k2 {
		t.Error("two generated private keys should not be identical")
	}
}

// --- worker.go ---

func TestAtomicCounter(t *testing.T) {
	c := &AtomicCounter{Value: 3}
	if c.Get() != 3 {
		t.Errorf("expected 3, got %d", c.Get())
	}
	c.Dec()
	if c.Get() != 2 {
		t.Errorf("expected 2 after Dec, got %d", c.Get())
	}
}

func TestCrunchWordMatch(t *testing.T) {
	opts := Options{Cores: 1, CaseSensitive: false}
	c := New(opts, 0)

	var matched []Pair
	buf := make([]byte, base64.StdEncoding.EncodedLen(KeySize))

	// Run crunch until we get one match for a short common prefix.
	// Base64 chars are a-z, A-Z, 0-9, +, / — single char prefix has 1/64 chance.
	c.WordMap["a"] = &AtomicCounter{Value: 1}
	for len(matched) == 0 {
		c.crunch(func(p Pair) { matched = append(matched, p) }, buf)
	}

	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
	if !strings.HasPrefix(strings.ToLower(matched[0].Public), "a") {
		t.Errorf("public key %q does not start with 'a'", matched[0].Public)
	}
}

func TestCrunchCaseSensitive(t *testing.T) {
	opts := Options{Cores: 1, CaseSensitive: true}
	c := New(opts, 0)

	var matched []Pair
	buf := make([]byte, base64.StdEncoding.EncodedLen(KeySize))

	c.WordMap["A"] = &AtomicCounter{Value: 1}
	for len(matched) == 0 {
		c.crunch(func(p Pair) { matched = append(matched, p) }, buf)
	}

	if !strings.HasPrefix(matched[0].Public, "A") {
		t.Errorf("public key %q does not start with 'A' (case-sensitive)", matched[0].Public)
	}
}

func TestCrunchRegexpMatch(t *testing.T) {
	opts := Options{Cores: 1, CaseSensitive: false}
	c := New(opts, 0)

	var matched []Pair
	buf := make([]byte, base64.StdEncoding.EncodedLen(KeySize))

	re := regexp.MustCompile(`(?i)^[ab]`)
	c.RegexpMap[re] = &AtomicCounter{Value: 1}
	for len(matched) == 0 {
		c.crunch(func(p Pair) { matched = append(matched, p) }, buf)
	}

	pub := strings.ToLower(matched[0].Public)
	if pub[0] != 'a' && pub[0] != 'b' {
		t.Errorf("public key %q does not match expected pattern", matched[0].Public)
	}
}

func TestCrunchCounterExhausted(t *testing.T) {
	opts := Options{Cores: 1, CaseSensitive: false}
	c := New(opts, 0)

	buf := make([]byte, base64.StdEncoding.EncodedLen(KeySize))
	c.WordMap["a"] = &AtomicCounter{Value: 0} // already exhausted

	var called int
	// Run a few iterations; counter is 0 so completed=true on first call
	for i := 0; i < 5; i++ {
		c.crunch(func(Pair) { called++ }, buf)
	}
	if called != 0 {
		t.Errorf("expected no matches when counter exhausted, got %d", called)
	}
}

func TestFindWordMatch(t *testing.T) {
	opts := Options{Cores: 2, CaseSensitive: false}
	c := New(opts, 0)
	c.WordMap["a"] = &AtomicCounter{Value: 1}

	var results []Pair
	c.Find(func(p Pair) { results = append(results, p) })

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !strings.HasPrefix(strings.ToLower(results[0].Public), "a") {
		t.Errorf("unexpected public key: %s", results[0].Public)
	}
}

func TestCollectToSlice(t *testing.T) {
	opts := Options{Cores: 2, CaseSensitive: false}
	c := New(opts, 0)
	c.WordMap["a"] = &AtomicCounter{Value: 2}

	results := c.CollectToSlice()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !strings.HasPrefix(strings.ToLower(r.Public), "a") {
			t.Errorf("unexpected public key: %s", r.Public)
		}
	}
}

func TestFindTimeout(t *testing.T) {
	opts := Options{Cores: 1, CaseSensitive: false}
	// Use a prefix that will never match
	c := New(opts, 100*time.Millisecond)
	c.WordMap["aaaaaaaaaa"] = &AtomicCounter{Value: 1}

	var results []Pair
	c.Find(func(p Pair) { results = append(results, p) })

	// Should have timed out with no results
	if len(results) != 0 {
		t.Errorf("expected no results after timeout, got %d", len(results))
	}
}

// --- utils.go ---

func TestIsValidSearch(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc123", true},
		{"ABC", true},
		{"a/b+c", true},
		{"abc!", false},
		{"abc ", false},
		{"abc@", false},
	}
	for _, tt := range tests {
		got := IsValidSearch(tt.input)
		if got != tt.valid {
			t.Errorf("IsValidSearch(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestIsRegex(t *testing.T) {
	tests := []struct {
		input   string
		isRegex bool
	}{
		{"abc", false},
		{"abc+", false}, // + is valid in keys, not a regex trigger
		{"abc*", true},
		{"^abc", true},
		{"abc$", true},
		{"[abc]", true},
		{"abc.def", true},
	}
	for _, tt := range tests {
		got := IsRegex(tt.input)
		if got != tt.isRegex {
			t.Errorf("IsRegex(%q) = %v, want %v", tt.input, got, tt.isRegex)
		}
	}
}

func TestIsValidRegex(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"^abc", false},
		{"abc$", true},  // ends with $ but no = before it
		{"abc=$", true}, // = is rejected by IsValidSearch after stripping $
		{"a.^b", true},  // ^ not at start
		{`\a`, true},    // will never match
	}
	for _, tt := range tests {
		got := IsValidRegex(tt.input)
		hasErr := got != ""
		if hasErr != tt.wantErr {
			t.Errorf("IsValidRegex(%q): got error=%v (%q), wantErr=%v", tt.input, hasErr, got, tt.wantErr)
		}
	}
}

func TestPlural(t *testing.T) {
	if Plural("key", 1) != "key" {
		t.Error("expected singular")
	}
	if Plural("key", 0) != "keys" {
		t.Error("expected plural for 0")
	}
	if Plural("key", 2) != "keys" {
		t.Error("expected plural for 2")
	}
}

func TestNumberFormat(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1000000, "1,000,000"},
		{-1000, "-1,000"},
	}
	for _, tt := range tests {
		got := NumberFormat(tt.n)
		if got != tt.want {
			t.Errorf("NumberFormat(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{5 * time.Second, "5 seconds"},
		{1 * time.Second, "1 second"},
		{90 * time.Second, "1 minute"},
		{2 * time.Minute, "2 minutes"},
		{2 * time.Hour, "2 hours, 0 minutes"},
		{48 * time.Hour, "2 days, 0 hours"},
	}
	for _, tt := range tests {
		got := HumanizeDuration(tt.d)
		if got != tt.want {
			t.Errorf("HumanizeDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestCalculateProbability(t *testing.T) {
	// Case-insensitive: alpha chars have higher probability (lower denominator)
	pInsensitive := CalculateProbability("a", false)
	pSensitive := CalculateProbability("a", true)
	if pInsensitive >= pSensitive {
		t.Errorf("case-insensitive probability (%d) should be lower than case-sensitive (%d)", pInsensitive, pSensitive)
	}

	// Longer prefix should have higher probability value (lower chance = higher number)
	p1 := CalculateProbability("a", false)
	p2 := CalculateProbability("ab", false)
	if p2 <= p1 {
		t.Errorf("two-char probability (%d) should exceed one-char (%d)", p2, p1)
	}
}
