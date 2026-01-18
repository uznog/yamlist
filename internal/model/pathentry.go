package model

// PathEntry is a flattened entry in the search index
// It provides fast lookup and fuzzy matching capabilities
type PathEntry struct {
	// Path is the full path to this node
	Path *Path

	// DisplayString is the pre-computed display string for fuzzy matching
	DisplayString string

	// Node is a reference to the actual node
	Node *Node
}

// NewPathEntry creates a new path entry from a node
func NewPathEntry(node *Node) *PathEntry {
	return &PathEntry{
		Path:          node.Path,
		DisplayString: node.Path.DisplayString(),
		Node:          node,
	}
}

// PathIndex is a collection of path entries for searching
type PathIndex struct {
	entries []*PathEntry
}

// NewPathIndex creates a new empty path index
func NewPathIndex() *PathIndex {
	return &PathIndex{
		entries: make([]*PathEntry, 0),
	}
}

// Add adds a path entry to the index
func (idx *PathIndex) Add(entry *PathEntry) {
	idx.entries = append(idx.entries, entry)
}

// Entries returns all entries in the index
func (idx *PathIndex) Entries() []*PathEntry {
	return idx.entries
}

// Len returns the number of entries in the index
func (idx *PathIndex) Len() int {
	return len(idx.entries)
}

// DisplayStrings returns all display strings for fuzzy matching
func (idx *PathIndex) DisplayStrings() []string {
	result := make([]string, len(idx.entries))
	for i, entry := range idx.entries {
		result[i] = entry.DisplayString
	}
	return result
}

// EntryAt returns the entry at the given index
func (idx *PathIndex) EntryAt(i int) *PathEntry {
	if i < 0 || i >= len(idx.entries) {
		return nil
	}
	return idx.entries[i]
}
