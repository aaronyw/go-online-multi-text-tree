package suffixtree

import (
	"fmt"
	"testing"
)

func BenchmarkUkkonenBuild(b *testing.B) {
	m := NewMultxtree(true)
	texts := []string{"banana", "mississippi", "abcdefghij"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newM := NewMultxtree(true)
		for _, t := range texts {
			newM.Read(t)
		}
	}
	_ = m
}

func BenchmarkUkkonenSearch(b *testing.B) {
	m := NewMultxtree(true)
	for _, t := range []string{"banana", "mississippi", "abcdefghij"} {
		m.Read(t)
	}
	queries := []string{"ana", "issi", "cde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Search(queries[i%len(queries)])
	}
}

func BenchmarkUkkonenLargeBuild(b *testing.B) {
	texts := make([]string, 50)
	for i := range texts {
		texts[i] = fmt.Sprintf("user-%d-email-%d@example.com", i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := NewMultxtree(true)
		for _, t := range texts {
			m.Read(t)
		}
		_ = m
	}
}
