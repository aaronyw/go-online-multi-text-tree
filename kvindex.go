package suffixtree

// KVIndex is a key-value store backed by a Multxtree suffix tree.
// It supports substring search over values, returning the keys whose values
// contain the query. Deletion is not supported; rebuild to remove entries.
type KVIndex struct {
	m     *Multxtree
	idKey map[int]string
}

// NewKVIndex returns an empty KVIndex.
func NewKVIndex() *KVIndex {
	return &KVIndex{
		m:     NewMultxtree(true),
		idKey: make(map[int]string),
	}
}

// Insert adds key with value to the index. The next available text ID is
// assigned to key so that Search can map matches back to keys.
func (k *KVIndex) Insert(key, value string) error {
	nextID := k.m.ID + 1
	if err := k.m.Read(value); err != nil {
		return err
	}
	k.idKey[nextID] = key
	return nil
}

// Search returns the deduplicated list of keys whose values contain substring.
func (k *KVIndex) Search(substring string) []string {
	hitIDs := k.m.searchIDs(substring)
	if len(hitIDs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(hitIDs))
	for id := range hitIDs {
		if key, ok := k.idKey[id]; ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// Len returns the number of entries in the index.
func (k *KVIndex) Len() int { return k.m.ID }
