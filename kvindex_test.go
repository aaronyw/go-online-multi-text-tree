package suffixtree

import (
	"sort"
	"testing"
)

func TestKVIndexBasic(t *testing.T) {
	idx := NewKVIndex()
	if err := idx.Insert("doc1", "hello world"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Insert("doc2", "goodbye world"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Insert("doc3", "hello friend"); err != nil {
		t.Fatal(err)
	}

	if got := idx.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	got := idx.Search("hello")
	sort.Strings(got)
	want := []string{"doc1", "doc3"}
	if len(got) != len(want) {
		t.Fatalf("Search(\"hello\") = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Search(\"hello\")[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	got = idx.Search("world")
	sort.Strings(got)
	want = []string{"doc1", "doc2"}
	if len(got) != len(want) {
		t.Fatalf("Search(\"world\") = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Search(\"world\")[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKVIndexMultipleMatches(t *testing.T) {
	idx := NewKVIndex()
	// "go" appears twice in the value; the key should appear once in results.
	if err := idx.Insert("lang", "go is great and go is fast"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Insert("other", "rust is fast"); err != nil {
		t.Fatal(err)
	}

	got := idx.Search("go")
	if len(got) != 1 || got[0] != "lang" {
		t.Errorf("Search(\"go\") = %v, want [lang]", got)
	}

	got = idx.Search("fast")
	sort.Strings(got)
	want := []string{"lang", "other"}
	if len(got) != len(want) {
		t.Fatalf("Search(\"fast\") = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Search(\"fast\")[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKVIndexNoMatch(t *testing.T) {
	idx := NewKVIndex()
	if err := idx.Insert("a", "apple"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Insert("b", "banana"); err != nil {
		t.Fatal(err)
	}

	got := idx.Search("cherry")
	if len(got) != 0 {
		t.Errorf("Search(\"cherry\") = %v, want []", got)
	}
}
