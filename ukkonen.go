package suffixtree

// Ukkonen is an online suffix tree built with Ukkonen's algorithm. It embeds
// *Radix and reuses its Search/Leaves/Display machinery, overriding only the
// insertion strategy (push) to maintain a true suffix tree rather than a radix
// tree.
//
// The active-point bookkeeping mirrors the canonical description at
// https://stackoverflow.com/questions/9452701/ukkonens-suffix-tree-algorithm-in-plain-english
//
// Hot-path note: dive and applySuffixRules are implemented as regular methods
// (not closures) to avoid per-character heap allocation. The constructor
// method zeroes a small per-call struct rather than allocating closures.
type Ukkonen struct {
	*Radix
	Debug bool

	Anode    int    // active node
	Aedge    int    // active edge (index into Text; 0 means "unset, use i")
	Alen     int    // active length
	Ext      int    // extent pointer (right edge of the implicit tree)
	Reminder int    // number of suffixes still to be inserted
	Prenode  [2]int // {last internal node created, Ext when it was created}

	// per-step state (zeroed each constructor call to avoid closure allocation)
	stepNewNode int
}

// NewUkkonen returns an empty suffix tree using DefaultETX as its delimiter.
func NewUkkonen() *Ukkonen {
	return &Ukkonen{Radix: NewRadix()}
}

// Read appends text (terminated by end, or the tree's ETX when end == 0) and
// extends the suffix tree one rune at a time.
func (u *Ukkonen) Read(text string, end rune) error {
	i := u.appendText(text, end)
	u.push(i)
	return nil
}

// push extends the tree to cover every suffix that begins at or after i.
func (u *Ukkonen) push(i int) {
	needSearch := true
	for i < len(u.Text) {
		needSearch, i = u.constructor(i, needSearch)
	}
}

// constructor performs one Ukkonen step for the rune at position i.
//
// In Go, local closures that capture variables escape to the heap.
// To avoid per-character allocation, diveExtend and applyRules are plain
// methods that read/write u.stepNewNode instead of capturing a local.
func (u *Ukkonen) constructor(i int, flag bool) (bool, int) {
	u.stepNewNode = 0

	u.Ext = max(u.Ext, i)
	if flag {
		res := u.search(u.Text[i:u.Ext], 0)
		u.Anode = res.Node
		u.Alen = res.PtrEdge
		if res.Found && u.Alen != 0 {
			u.Aedge = u.Ext + res.Edge
		} else {
			u.Aedge = u.Ext + res.Edge + 1
		}
	}

	if u.diveExtend(u.Ext) {
		u.Ext = min(len(u.Text)-1, u.Ext+1)
		return false, i
	}

	if u.Alen != 0 {
		node := NewNode()
		u.stepNewNode = len(u.Nodes)
		aedge := u.Nodes[u.Anode].Edges[u.Text[u.Aedge]]
		split := aedge.Start + u.Alen
		node.Edges[u.Text[split]] = &Edge{Src: u.stepNewNode, Start: split, End: aedge.End, Des: aedge.Des}
		aedge.End = split
		aedge.Des = u.stepNewNode
		node.Edges[u.Text[u.Ext]] = &Edge{Src: u.stepNewNode, Start: u.Ext, End: EndSentinel, Des: DesSentinel}
		u.Nodes = append(u.Nodes, node)
		u.Reminder--
	} else {
		u.Nodes[u.Anode].Edges[u.Text[u.Ext]] = &Edge{Src: u.Anode, Start: u.Ext, End: EndSentinel, Des: DesSentinel}
		if u.Anode != 0 {
			u.Reminder--
		}
	}
	return u.applyRules(), i + 1
}

// diveExtend walks the active point down existing edges trying to find the
// rune at position j. It returns true when the rune already exists (implicit
// extension / Ukkonen rule 3).
func (u *Ukkonen) diveExtend(j int) bool {
	if u.Aedge == 0 {
		u.Aedge = j
	}
	aedge, ok := u.Nodes[u.Anode].Edges[u.Text[u.Aedge]]
	if !ok {
		return false
	}
	if aedge.Des != DesSentinel && u.Alen >= aedge.End-aedge.Start {
		u.Anode = aedge.Des
		u.Alen -= aedge.End - aedge.Start
		u.Aedge = j - u.Alen
		return u.diveExtend(j)
	}
	if u.Text[j] == u.Text[aedge.Start+u.Alen] {
		u.Alen++
		u.Reminder++
		return true
	}
	return false
}

// applyRules implements Ukkonen rules 1/2/3 after an explicit edge insertion.
// Returns true if the next step needs a fresh search from the root.
func (u *Ukkonen) applyRules() bool {
	if u.stepNewNode != 0 {
		// Inline child check — avoids allocating 3 slices via Children().
		isChild := false
		for _, e := range u.Nodes[u.Prenode[0]].Edges {
			if e.Des == u.stepNewNode {
				isChild = true
				break
			}
		}
		if u.Prenode[1] == u.Ext && !isChild {
			u.Nodes[u.Prenode[0]].SuffixLink = u.stepNewNode
		}
		u.Prenode = [2]int{u.stepNewNode, u.Ext}
	}
	link := u.Nodes[u.Anode].SuffixLink
	if u.Anode != 0 {
		u.Anode = link
		if link == 0 && u.Reminder != 0 {
			return true
		}
	} else {
		u.Aedge = min(len(u.Text)-1, u.Aedge+1)
		u.Alen = max(u.Alen-1, 0)
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
