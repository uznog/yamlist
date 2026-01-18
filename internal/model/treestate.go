package model

// TreeState holds the current state of the tree view
type TreeState struct {
	// Expanded tracks which paths are expanded
	// Key is the path string representation
	Expanded map[string]bool

	// SelectedIndex is the index of the currently selected row in visible rows
	SelectedIndex int

	// SelectedNode is a reference to the currently selected node
	SelectedNode *Node

	// VisibleRows is the list of currently visible rows
	VisibleRows []*VisibleRow

	// Root is the root node of the tree
	Root *Node

	// ScrollOffset is the number of rows scrolled from the top
	ScrollOffset int
}

// NewTreeState creates a new tree state with the given root
func NewTreeState(root *Node) *TreeState {
	return &TreeState{
		Expanded:      make(map[string]bool),
		SelectedIndex: 0,
		SelectedNode:  root,
		VisibleRows:   make([]*VisibleRow, 0),
		Root:          root,
		ScrollOffset:  0,
	}
}

// IsExpanded returns true if the node at the given path is expanded
func (ts *TreeState) IsExpanded(path *Path) bool {
	if path == nil {
		return true // root is always expanded
	}
	return ts.Expanded[path.String()]
}

// SetExpanded sets the expanded state for a path
func (ts *TreeState) SetExpanded(path *Path, expanded bool) {
	if path == nil {
		return // can't collapse root
	}
	pathStr := path.String()
	if expanded {
		ts.Expanded[pathStr] = true
	} else {
		delete(ts.Expanded, pathStr)
	}
}

// ToggleExpanded toggles the expanded state for a path
func (ts *TreeState) ToggleExpanded(path *Path) bool {
	if path == nil {
		return true
	}
	pathStr := path.String()
	if ts.Expanded[pathStr] {
		delete(ts.Expanded, pathStr)
		return false
	}
	ts.Expanded[pathStr] = true
	return true
}

// ExpandAll expands all expandable nodes
func (ts *TreeState) ExpandAll() {
	ts.expandAllRecursive(ts.Root)
}

func (ts *TreeState) expandAllRecursive(node *Node) {
	if node == nil {
		return
	}
	if node.IsExpandable() && node.HasChildren() {
		ts.SetExpanded(node.Path, true)
	}
	for _, child := range node.Children {
		ts.expandAllRecursive(child)
	}
}

// CollapseAll collapses all nodes
func (ts *TreeState) CollapseAll() {
	ts.Expanded = make(map[string]bool)
}

// ExpandToNode expands all ancestors of the given node
func (ts *TreeState) ExpandToNode(node *Node) {
	if node == nil {
		return
	}
	current := node.Parent
	for current != nil {
		ts.SetExpanded(current.Path, true)
		current = current.Parent
	}
}

// MoveSelection moves the selection by delta rows
// Returns true if the selection changed
func (ts *TreeState) MoveSelection(delta int) bool {
	if len(ts.VisibleRows) == 0 {
		return false
	}

	newIndex := ts.SelectedIndex + delta
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(ts.VisibleRows) {
		newIndex = len(ts.VisibleRows) - 1
	}

	if newIndex == ts.SelectedIndex {
		return false
	}

	ts.SelectedIndex = newIndex
	ts.SelectedNode = ts.VisibleRows[newIndex].Node
	return true
}

// SelectNode selects a specific node
// Returns true if the node was found and selected
func (ts *TreeState) SelectNode(node *Node) bool {
	for i, row := range ts.VisibleRows {
		if row.Node == node {
			ts.SelectedIndex = i
			ts.SelectedNode = node
			return true
		}
	}
	return false
}

// SelectByPath finds and selects a node by path
func (ts *TreeState) SelectByPath(path *Path) bool {
	for i, row := range ts.VisibleRows {
		if row.Node.Path.Equal(path) {
			ts.SelectedIndex = i
			ts.SelectedNode = row.Node
			return true
		}
	}
	return false
}

// GetSelectedRow returns the currently selected row
func (ts *TreeState) GetSelectedRow() *VisibleRow {
	if ts.SelectedIndex < 0 || ts.SelectedIndex >= len(ts.VisibleRows) {
		return nil
	}
	return ts.VisibleRows[ts.SelectedIndex]
}
