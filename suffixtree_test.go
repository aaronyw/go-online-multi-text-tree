package suffixtree

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// sortedPositions returns a copy of the result map with every position slice
// sorted, so comparisons are independent of Go's random map iteration order.
func sortedPositions(m map[int][]int) map[int][]int {
	out := make(map[int][]int, len(m))
	for k, v := range m {
		c := append([]int(nil), v...)
		sort.Ints(c)
		out[k] = c
	}
	return out
}

func assertSearch(t *testing.T, got, want map[int][]int) {
	t.Helper()
	if !reflect.DeepEqual(sortedPositions(got), want) {
		t.Errorf("search mismatch:\n got  %v\n want %v", sortedPositions(got), want)
	}
}

// fmtSearch reproduces the Demo.ipynb formatting: it highlights each match by
// wrapping it in |term| and returns one reconstructed string per matched text.
func fmtSearch(tr *Multxtree, s string) map[int]string {
	out := map[int]string{}
	for i, pos := range tr.Search(s) {
		segs := tr.Extractxt(i, len([]rune(s)), pos)
		out[i] = strings.Join(segs, "|"+s+"|")
	}
	return out
}

func mustRead(t *testing.T, tr *Multxtree, texts ...string) {
	t.Helper()
	for _, txt := range texts {
		if err := tr.Read(txt); err != nil {
			t.Fatalf("Read(%q): %v", txt, err)
		}
	}
}

// --- Basic Radix prefix search ---------------------------------------------

func TestRadixPrefixSearch(t *testing.T) {
	r := NewMultxtree(false)
	mustRead(t, r, "abc", "abx", "abcd")

	// "ab" is a prefix of all three texts.
	assertSearch(t, r.Search("ab"), map[int][]int{1: {2}, 2: {6}, 3: {10}})
	// "abc" is a prefix of texts 1 and 3 only.
	assertSearch(t, r.Search("abc"), map[int][]int{1: {3}, 3: {11}})
	// A radix tree only matches prefixes, so an interior substring misses.
	assertSearch(t, r.Search("bc"), map[int][]int{})
	// A string that is nowhere in the tree.
	assertSearch(t, r.Search("zz"), map[int][]int{})

	// Highlighted reconstruction, mirroring the original demo output.
	if got, want := fmtSearch(r, "abc"), map[int]string{1: "|abc|", 3: "|abc|d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("fmtSearch(abc) = %v, want %v", got, want)
	}
}

func TestRadixLeaves(t *testing.T) {
	r := NewRadix()
	if err := r.Read("cat", 0); err != nil {
		t.Fatal(err)
	}
	if err := r.Read("car", 0); err != nil {
		t.Fatal(err)
	}
	leaves := r.Leaves()
	// Every full string stored ends in the '$' terminator. Both words share the
	// "ca" prefix, so the leaves below it are "t$" and "r$".
	joined := strings.Join(leaves, ",")
	if !strings.Contains(joined, "t$") || !strings.Contains(joined, "r$") {
		t.Errorf("Leaves() = %v, expected leaves for t$ and r$", leaves)
	}
}

// --- Ukkonen substring search ----------------------------------------------

func TestUkkonenSubstringSearch(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "abc", "abx", "abcd")

	// Unlike the radix tree, the suffix tree finds interior substrings.
	assertSearch(t, u.Search("bc"), map[int][]int{1: {3}, 3: {11}})
	assertSearch(t, u.Search("abc"), map[int][]int{1: {3}, 3: {11}})
	assertSearch(t, u.Search("a"), map[int][]int{1: {1}, 2: {5}, 3: {9}})

	if got, want := fmtSearch(u, "bc"), map[int]string{1: "a|bc|", 3: "a|bc|d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("fmtSearch(bc) = %v, want %v", got, want)
	}
}

func TestUkkonenBanana(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "banana")

	// Classic suffix-tree example. Positions are match end offsets.
	assertSearch(t, u.Search("ana"), map[int][]int{1: {4, 6}})
	assertSearch(t, u.Search("na"), map[int][]int{1: {4, 6}})
	assertSearch(t, u.Search("an"), map[int][]int{1: {3, 5}})
	assertSearch(t, u.Search("banana"), map[int][]int{1: {6}})
	assertSearch(t, u.Search("xyz"), map[int][]int{})
}

// --- Multiple text insertion with delimiter separation ---------------------

func TestDelimiterSeparation(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "abc", "xyz")

	// Each text gets a unique astral-plane delimiter so suffixes never bleed
	// across texts: searching one text's content must not match the other.
	assertSearch(t, u.Search("abc"), map[int][]int{1: {3}})
	assertSearch(t, u.Search("xyz"), map[int][]int{2: {7}})

	// The delimiters are distinct runes >= 0x10000.
	text := u.Tree.TextRunes()
	d1, d2 := text[3], text[7]
	if d1 < 0x10000 || d2 < 0x10000 {
		t.Errorf("delimiters not in astral plane: %#U %#U", d1, d2)
	}
	if d1 == d2 {
		t.Errorf("delimiters must be unique, both = %#U", d1)
	}

	// A cross-text "substring" formed only by spanning a delimiter must miss.
	assertSearch(t, u.Search("cx"), map[int][]int{})
}

func TestExtractxtReconstruction(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "abcabc")

	res := u.Search("abc") // matches at ends 3 and 6
	segs := u.Extractxt(1, 3, res[1])
	got := strings.Join(segs, "|abc|")
	if got != "|abc||abc|" {
		t.Errorf("reconstruction = %q, want %q", got, "|abc||abc|")
	}
}

// --- Edge cases ------------------------------------------------------------

func TestEmptyString(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "")
	// An empty text is just its terminator; the empty query matches at 0.
	assertSearch(t, u.Search(""), map[int][]int{1: {0}})
	// A non-empty query cannot match an empty text.
	assertSearch(t, u.Search("a"), map[int][]int{})
}

func TestSingleChar(t *testing.T) {
	for _, ukk := range []bool{false, true} {
		tr := NewMultxtree(ukk)
		mustRead(t, tr, "a")
		assertSearch(t, tr.Search("a"), map[int][]int{1: {1}})
		assertSearch(t, tr.Search("b"), map[int][]int{})
	}
}

func TestDuplicateStrings(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "abc", "abc")
	// Identical texts are distinguished by their delimiters and reported
	// separately (match end offset 3 in text 1, offset 7 in text 2).
	assertSearch(t, u.Search("abc"), map[int][]int{1: {3}, 2: {7}})

	r := NewMultxtree(false)
	mustRead(t, r, "abc", "abc")
	assertSearch(t, r.Search("abc"), map[int][]int{1: {3}, 2: {7}})
}

func TestUnicodeContent(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "héllo", "wörld")
	// Multi-byte runes are indexed by code point, so the match offset is in
	// runes, not bytes.
	assertSearch(t, u.Search("llo"), map[int][]int{1: {5}})
	assertSearch(t, u.Search("ör"), map[int][]int{2: {9}})
}

func TestDisplayDoesNotPanic(t *testing.T) {
	u := NewMultxtree(true)
	mustRead(t, u, "banana")
	if s := u.Tree.(*Ukkonen).String(); s == "" {
		t.Error("String() returned empty output")
	}
}
