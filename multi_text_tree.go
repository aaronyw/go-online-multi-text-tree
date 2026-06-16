package suffixtree

import (
	"fmt"
	"sort"
)

// DefaultShift is the first delimiter code point. Texts are separated by unique
// runes starting at this value (0x10000, the first astral-plane code point),
// which is why the tree stores text as []rune.
const DefaultShift = 0x10000

// messageLimit bounds the number of distinct texts (and is reused as the
// default cap on the number of matching texts returned by Search).
const messageLimit = 1 << 21

// Tree is the behaviour shared by Radix and Ukkonen that Multxtree relies on.
type Tree interface {
	Read(text string, end rune) error
	Search(text []rune) SearchResult
	AllNodes() []*Node
	TextRunes() []rune
}

// Multxtree stores several texts in one tree, separating them with unique
// Unicode delimiters so that suffixes/words from different texts never merge.
type Multxtree struct {
	Shift   int  // first delimiter code point
	ID      int  // number of texts inserted so far
	Ukkonen bool // true => suffix tree, false => radix (prefix) tree
	Tree    Tree
	Index   []int // Index[k] = end offset (exclusive) of text k in Tree's text
}

// NewMultxtree builds a multi-text tree. Pass true for a Ukkonen suffix tree
// (substring search) or false for a Radix tree (prefix search).
func NewMultxtree(ukkonen bool) *Multxtree {
	var t Tree
	if ukkonen {
		t = NewUkkonen()
	} else {
		t = NewRadix()
	}
	return &Multxtree{Shift: DefaultShift, Ukkonen: ukkonen, Tree: t, Index: []int{0}}
}

// Read inserts a new text, terminating it with a per-text unique delimiter.
func (m *Multxtree) Read(text string) error {
	m.ID++
	if m.ID == messageLimit-m.Shift {
		return fmt.Errorf("reached the message limit %d", messageLimit)
	}
	if err := m.Tree.Read(text, rune(m.ID+m.Shift)); err != nil {
		return err
	}
	m.Index = append(m.Index, len(m.Tree.TextRunes()))
	return nil
}

// Search finds every occurrence of text and returns a map from text ID
// (1-based, matching the insertion order) to the start positions of the matches
// within the concatenated text. An empty map means no match.
func (m *Multxtree) Search(text string) map[int][]int {
	return m.SearchLimit(text, messageLimit)
}

// SearchLimit is Search with an explicit cap on the number of distinct texts to
// report.
func (m *Multxtree) SearchLimit(text string, limit int) map[int][]int {
	query := []rune(text)
	res := m.Tree.Search(query)
	r := map[int][]int{}
	if !res.Found {
		return r
	}
	nodes := m.Tree.AllNodes()

	n := res.Node
	shift := 0
	if res.PtrEdge > 0 {
		// The query ended in the middle of an edge.
		key := query[len(query)+res.Edge] // query[res.Edge] with Python negative indexing
		edge := nodes[n].Edges[key]
		if !edge.IsLeaf() {
			shift = edge.End - edge.Start - res.PtrEdge
			n = edge.Des
		} else {
			pos := edge.Start + res.PtrEdge
			return map[int][]int{m.bisect(pos): {pos}}
		}
	}

	// Breadth-first walk of the matched subtree, accumulating edge lengths into
	// shift so that each leaf's start position resolves to where the matched
	// substring began.
	type qItem struct{ node, shift int }
	q := []qItem{{n, shift}}
	for i := 0; i < len(q) && len(r) <= limit; i++ {
		branches, branchLen, leaves := nodes[q[i].node].Children()
		for j, b := range branches {
			q = append(q, qItem{node: b, shift: branchLen[j] + q[i].shift})
		}
		for _, leaf := range leaves {
			pos := leaf - q[i].shift
			k := m.bisect(pos)
			r[k] = append(r[k], pos)
		}
	}
	return r
}

// Extractxt returns the non-highlighted segments of text id i, split around the
// highlight positions. hlLen is the highlight (match) length and hlPos the
// highlight end positions. Joining the returned segments with the highlighted
// term reconstructs the original text.
func (m *Multxtree) Extractxt(i, hlLen int, hlPos []int) []string {
	text := m.Tree.TextRunes()
	positions := append([]int(nil), hlPos...)
	sort.Ints(positions)

	s := m.Index[i-1]
	var r []string
	for _, h := range positions {
		r = append(r, string(text[s:h-hlLen]))
		s = h
	}
	// Trailing segment, dropping the delimiter rune at the end of the text.
	r = append(r, string(text[s:m.Index[i]-1]))
	return r
}

// bisect mirrors Python's bisect.bisect_right over the Index slice, mapping an
// absolute position to its 1-based text ID.
func (m *Multxtree) bisect(pos int) int {
	lo, hi := 0, len(m.Index)
	for lo < hi {
		mid := (lo + hi) / 2
		if pos < m.Index[mid] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}
