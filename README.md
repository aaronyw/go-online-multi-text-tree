# go-online-multi-text-tree

Go port of [Online-Mulit-Text-Tree-Ukkonen-Algorithms](https://github.com/aaronyw/Online-Mulit-Text-Tree-Ukkonen-Algorithms).

Package `suffixtree` provides a compressed prefix tree (Radix), a linear-time suffix tree (Ukkonen), and a multi-text wrapper that separates texts with unique Unicode delimiters. An added `KVIndex` layer wraps the suffix tree for value-substring → key lookup.

## Import

```go
import "github.com/aaronyw/go-online-multi-text-tree"
```

## Quick Start

### Substring Search (Suffix Tree)

```go
m := suffixtree.NewMultxtree(true) // true = Ukkonen suffix tree
m.Read("banana")
m.Read("mississippi")

results := m.Search("ana")
for id, positions := range results {
    fmt.Printf("Text %d: matches at %v\n", id, positions)
}
```

### Prefix Search (Radix Tree)

```go
r := suffixtree.NewMultxtree(false) // false = radix tree
r.Read("hello")
r.Read("world")
```

### Key / Value Index

`KVIndex` maps a key to a value string and lets you find keys by value substring — a common pattern for search-as-you-type, tag-based lookup, or inverted-index-light use cases.

```go
idx := suffixtree.NewKVIndex()
idx.Insert("user1", "John Smith")
idx.Insert("user2", "Jane Doe")
idx.Insert("user3", "John Appleseed")

// Find all keys whose value contains "John"
keys := idx.Search("John")        // → ["user1", "user3"]

// No match returns an empty slice
keys = idx.Search("xyz")          // → []
```

Under the hood each value is indexed in the suffix tree. Search traverses O(len(query)) steps regardless of index size, then deduplicates keys that match multiple times in the same value (e.g. `"go is great and go is fast"` searching `"go"` → `["lang"]` once, not twice).

## API

### Radix Tree

| Function | Description |
|----------|-------------|
| `NewRadix()` | Empty radix tree |
| `Read(text, end)` | Insert text with delimiter |
| `Search(text)` | Walk tree, return match result |
| `Leaves()` | All stored strings, sorted by length |
| `Display(limit)` | Debug print (limit edge labels) |

### Ukkonen Suffix Tree

| Function | Description |
|----------|-------------|
| `NewUkkonen()` | Empty suffix tree |
| `Read(text, end)` | Insert text (linear time per char) |
| `Search(text)` | Prefix/suffix walk (inherits Radix) |

### Multi-Text Wrapper

| Function | Description |
|----------|-------------|
| `NewMultxtree(ukkonen)` | True = suffix tree, false = radix |
| `Read(text)` | Insert text with unique delimiter |
| `Search(text)` | Returns `map[textID][]position` |
| `SearchLimit(text, limit)` | Cap distinct texts returned |
| `Extractxt(id, hlLen, hlPos)` | Split text around highlights |

### KV Index

| Function | Description |
|----------|-------------|
| `NewKVIndex()` | Empty key-value index backed by suffix tree |
| `Insert(key, value)` | Index a key-value pair |
| `Search(substring)` | Return all keys whose value contains the substring |
| `Len()` | Number of indexed pairs |

## Performance (Raspberry Pi CM5, arm64, Go 1.24)

```
BenchmarkUkkonenBuild-4       105,000 ops    11,700 ns/op     5 kB     92 allocs
BenchmarkUkkonenSearch-4    2,240,000 ops       550 ns/op   400 B      5 allocs
BenchmarkUkkonenLargeBuild-4   3,060 ops   408,000 ns/op   165 kB  2,489 allocs

BenchmarkKVIndexInsert-4     104,000 ops    14,800 ns/op     4 kB     52 allocs
BenchmarkKVIndexSearch-4       510 ops  2,350,000 ns/op   464 kB     53 allocs
```

**Notes (test volumes):**
- `UkkonenBuild`: inserts 3 short texts (banana, mississippi, abcdefghij).
- `UkkonenLargeBuild`: inserts 50 medium texts.
- `UkkonenSearch`: runs queries `"ana"`, `"issi"`, `"cde"` against the built tree — ns/op because O(|query|) tree descent is very fast.
- `KVIndexInsert`: inserts one key-value pair per op.
- **`KVIndexSearch`**: runs 5 query patterns across a **10,000-entry** index. The sub-2.5 ms per search is dominated by suffix-tree leaf enumeration under the matching node. Tree descent itself stays O(|query|) — wall time scales with result set size, not index size. The internal `searchIDs` method avoids position storage, keeping allocations to just 53/op (down from 7,220 before optimisation).

## License

Same as original repository.
