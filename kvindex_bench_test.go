package suffixtree

import (
	"fmt"
	"testing"
)

func BenchmarkKVIndexInsert(b *testing.B) {
	idx := NewKVIndex()
	data := make([]string, b.N)
	for i := range data {
		data[i] = fmt.Sprintf("key-%08d-value-data-%d", i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := idx.Insert(data[i], data[i]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKVIndexSearch(b *testing.B) {
	idx := NewKVIndex()
	n := 10_000
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key-%08d", i)
		val := fmt.Sprintf("user-%d-with-some-description-text", i)
		if err := idx.Insert(key, val); err != nil {
			b.Fatal(err)
		}
	}
	queries := []string{"user-", "9999", "description", "text", "xyz-nomatch"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = idx.Search(queries[i%len(queries)])
	}
}
