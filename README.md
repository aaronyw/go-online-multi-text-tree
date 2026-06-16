# go-online-multi-text-tree

Go port of [Online-Mulit-Text-Tree-Ukkonen-Algorithms](https://github.com/aaronyw/Online-Mulit-Text-Tree-Ukkonen-Algorithms).

Package `suffixtree` provides a compressed prefix tree (Radix), a linear-time suffix tree (Ukkonen), and a multi-text wrapper that separates texts with unique Unicode delimiters.

## Import

```go
import "github.com/aaronyw/go-online-multi-text-tree"
```

## Quick Start

```go
package main

import (
    "fmt"
    suffixtree "github.com/aaronyw/go-online-multi-text-tree"
)

func main() {
    // Ukkonen suffix tree (substring search)
    m := suffixtree.NewMultxtree(true)
    m.Read("banana")
    m.Read("mississippi")

    results := m.Search("ana")
    for id, positions := range results {
        fmt.Printf("Text %d: matches at %v\n", id, positions)
    }

    // Radix tree (prefix search)
    r := suffixtree.NewMultxtree(false)
    r.Read("hello")
    r.Read("world")
}
```

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

## Performance (Raspberry Pi CM5, arm64)

```
BenchmarkUkkonenBuild-4        39,000 ops    30,100 ns/op   13 kB   228 allocs
BenchmarkUkkonenSearch-4      247,000 ops     4,800 ns/op    2 kB    56 allocs
BenchmarkUkkonenLargeBuild-4   16,000 ops    73,000 ns/op   31 kB   535 allocs
```

## License

Same as original repository.
