// Package suffixtree is an idiomatic Go port of the Python
// "Online-Multi-Text-Tree-Ukkonen-Algorithms" project.
//
// Repository: github.com/aaronyw/go-online-multi-text-tree
package suffixtree

import (
	"fmt"
	"strings"
)

// DefaultETX is the end-of-text marker used by a Radix tree when no explicit
// delimiter is supplied.
const DefaultETX = '$'

// Sentinel values for Edge.End and Edge.Des.
// EndSentinel means "open leaf edge" (conceptually extends to end of text).
// DesSentinel means "leaf edge" (no destination node).
const (
	EndSentinel = -1
	DesSentinel = -1
)

// Edge is a labelled edge between two nodes.
//
// The label is the run of runes Text[Start:End] where End is the exclusive end
// for an internal edge. End == EndSentinel marks an "open" leaf edge whose
// label conceptually extends to the current end of the text. Des == DesSentinel
// marks a leaf (the edge has no destination node).
//
// Sentinel values (-1) are safe because Start and End are always >= 0 for real
// positions, and node indices are always >= 0. This avoids the allocation and
// pointer-chasing overhead of *int used in the original Python port.
type Edge struct {
	Src   int // source node index
	Start int // inclusive start position of the label in the text
	End   int // exclusive end position; EndSentinel => open leaf edge
	Des   int // destination node index; DesSentinel => leaf
}

// IsLeaf reports whether the edge is an open leaf edge.
func (e *Edge) IsLeaf() bool { return e.End == EndSentinel }

// Node is a vertex of the tree.
type Node struct {
	Edges map[rune]*Edge
	// SuffixLink is the Ukkonen suffix link target node index. 0 means "no
	// link": node 0 is the root and is never the target of a suffix link, so 0
	// is an unambiguous sentinel here (unlike End/Des, this is not the falsy
	// bug).
	SuffixLink int
}

// NewNode returns an empty node.
func NewNode() *Node {
	return &Node{Edges: make(map[rune]*Edge)}
}

// Children partitions a node's edges into internal branches and leaves.
// It returns the destination node of each branch, the rune length of each
// branch's label (parallel to branches), and the start position of each leaf
// edge.
func (n *Node) Children() (branches, branchLength, leaves []int) {
	for _, e := range n.Edges {
		if e.Des != DesSentinel {
			branches = append(branches, e.Des)
			branchLength = append(branchLength, e.End-e.Start)
		} else {
			leaves = append(leaves, e.Start)
		}
	}
	return
}

// SearchResult is the outcome of a tree walk.
//
// The Edge and PtrText fields are negative offsets relative to the END of the
// text that was searched (mirroring the original Python implementation, which
// exploited negative indexing). They are converted to absolute positions at the
// call sites with helpers such as runeAt.
type SearchResult struct {
	Found   bool // whether the full query was matched
	Node    int  // node at/above the matched position
	Edge    int  // negative offset of the edge's first rune (-len(remaining))
	PtrEdge int  // how many runes were matched along the current edge
	PtrText int  // negative offset into the query of the unmatched remainder
}

// Radix is a compressed prefix (radix) tree supporting online insertion.
type Radix struct {
	Text  []rune
	Nodes []*Node
	ETX   rune // end-of-text delimiter used by Read when none is given
}

// NewRadix returns an empty Radix tree using DefaultETX as its delimiter.
func NewRadix() *Radix {
	return &Radix{Nodes: []*Node{NewNode()}, ETX: DefaultETX}
}

// AllNodes exposes the node slice (implements Tree).
func (r *Radix) AllNodes() []*Node { return r.Nodes }

// TextRunes exposes the underlying text (implements Tree).
func (r *Radix) TextRunes() []rune { return r.Text }

// Search walks the tree from the root looking for text (implements Tree).
func (r *Radix) Search(text []rune) SearchResult { return r.search(text, 0) }

// runeAt indexes Text the way Python indexes str, allowing negative offsets
// (counted from the end). It is used to convert the negative offsets returned by
// search into the rune they refer to.
func (r *Radix) runeAt(i int) rune {
	if i < 0 {
		return r.Text[len(r.Text)+i]
	}
	return r.Text[i]
}

// appendText appends text plus a delimiter and returns the start index of the
// freshly appended segment. An end of 0 selects the tree's ETX.
func (r *Radix) appendText(text string, end rune) int {
	i := len(r.Text)
	r.Text = append(r.Text, []rune(text)...)
	if end == 0 {
		end = r.ETX
	}
	r.Text = append(r.Text, end)
	return i
}

// Read appends text (terminated by end, or the tree's ETX when end == 0) and
// inserts the new suffix/word into the tree.
func (r *Radix) Read(text string, end rune) error {
	i := r.appendText(text, end)
	r.push(i)
	return nil
}

// push inserts the word that starts at position i (running to the end of the
// text) into the radix tree.
func (r *Radix) push(i int) {
	res := r.search(r.Text[i:], 0)
	n, e, p, pt := res.Node, res.Edge, res.PtrEdge, res.PtrText
	if p > 0 {
		// The new word diverges in the middle of an existing edge: split it.
		aedge := r.Nodes[n].Edges[r.runeAt(e)]
		node := NewNode()
		newIdx := len(r.Nodes)
		split := p + aedge.Start
		node.Edges[r.Text[split]] = &Edge{Src: newIdx, Start: split, End: aedge.End, Des: aedge.Des}
		leafStart := len(r.Text) + pt
		node.Edges[r.runeAt(pt)] = &Edge{Src: newIdx, Start: leafStart, End: EndSentinel, Des: DesSentinel}
		aedge.End = split
		aedge.Des = newIdx
		r.Nodes = append(r.Nodes, node)
	} else {
		leafStart := len(r.Text) + pt
		r.Nodes[n].Edges[r.runeAt(e)] = &Edge{Src: n, Start: leafStart, End: EndSentinel, Des: DesSentinel}
	}
}

// search matches text against the tree starting at node idx.
// Iterative version to avoid recursive call overhead per edge level.
func (r *Radix) search(text []rune, idx int) SearchResult {
	for {
		if len(text) == 0 {
			return SearchResult{Found: true, Node: idx, Edge: -1}
		}
		edge := -len(text)
		e, ok := r.Nodes[idx].Edges[text[0]]
		if !ok {
			return SearchResult{Found: false, Node: idx, Edge: edge, PtrText: -len(text)}
		}
		end := len(r.Text)
		if e.End != EndSentinel {
			end = e.End
		}
		i := 1
		for e.Start+i < end {
			if i == len(text) {
				return SearchResult{Found: true, Node: idx, Edge: edge, PtrEdge: i}
			} else if r.Text[e.Start+i] != text[i] {
				return SearchResult{Found: false, Node: idx, Edge: edge, PtrEdge: i, PtrText: i - len(text)}
			}
			i++
		}
		if e.Des == DesSentinel {
			panic("suffixtree: walked off a leaf edge during search")
		}
		text = text[i:]
		idx = e.Des
	}
}

// Leaves returns every full string stored in the tree, sorted by length.
func (r *Radix) Leaves() []string {
	if len(r.Text) == 0 {
		return nil
	}
	var dfs func(i int) []string
	dfs = func(i int) []string {
		var res []string
		for _, e := range r.Nodes[i].Edges {
			var rest []string
			if e.Des != DesSentinel {
				rest = dfs(e.Des)
			} else {
				rest = []string{""}
			}
			end := len(r.Text)
			if e.End != EndSentinel {
				end = e.End
			}
			label := string(r.Text[e.Start:end])
			for _, s := range rest {
				res = append(res, label+s)
			}
		}
		return res
	}
	res := dfs(0)
	for a := 1; a < len(res); a++ {
		for b := a; b > 0 && len(res[b-1]) > len(res[b]); b-- {
			res[b-1], res[b] = res[b], res[b-1]
		}
	}
	return res
}

// String renders the whole tree for debugging.
func (r *Radix) String() string {
	var sb strings.Builder
	r.display(&sb, 0, -1, "")
	return sb.String()
}

// Display renders the tree, truncating every edge label at the given absolute
// limit (use limit < 0 for no limit).
func (r *Radix) Display(limit int) string {
	var sb strings.Builder
	r.display(&sb, 0, limit, "")
	return sb.String()
}

func (r *Radix) display(sb *strings.Builder, i, limit int, pre string) {
	if pre != "" {
		fmt.Fprintln(sb, pre+"|")
	}
	if r.Nodes[i].SuffixLink != 0 {
		fmt.Fprintf(sb, "%sNODE %d\t------> node %d\n", pre, i, r.Nodes[i].SuffixLink)
	} else {
		fmt.Fprintf(sb, "%sNODE %d\n", pre, i)
	}
	for label, e := range r.Nodes[i].Edges {
		endPos := len(r.Text)
		if e.End != EndSentinel {
			endPos = e.End
		}
		if limit >= 0 && endPos > limit {
			endPos = limit
		}
		dst := ""
		if e.Des != DesSentinel {
			dst = fmt.Sprintf("\t==> node %d", e.Des)
		}
		endStr := "leaf"
		if e.End != EndSentinel {
			endStr = fmt.Sprintf("%d", e.End)
		}
		fmt.Fprintf(sb, "%s|___ %d edge %q: %d->%s\t[%s]%s\n",
			pre, e.Src, string(label), e.Start, endStr, string(r.Text[e.Start:endPos]), dst)
		if e.Des != DesSentinel {
			r.display(sb, e.Des, limit, pre+"|    ")
			fmt.Fprintln(sb, pre+"|    ")
		}
	}
	fmt.Fprintf(sb, "%s|___ end of node %d\n", pre, i)
}
