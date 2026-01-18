package model

// VisibleRow represents a single visible row in the tree view
// It contains all data needed for rendering
type VisibleRow struct {
	// Node is the underlying YAML node
	Node *Node

	// Depth is the indentation level
	Depth int

	// IsExpanded indicates if this node is currently expanded
	IsExpanded bool

	// IsExpandable indicates if this node can be expanded
	IsExpandable bool

	// IsSelected indicates if this row is currently selected
	IsSelected bool

	// HasChildren indicates if this node has children
	HasChildren bool

	// ChildCount is the number of direct children
	ChildCount int

	// Index is the row index in the visible rows list
	Index int
}

// NewVisibleRow creates a visible row from a node
func NewVisibleRow(node *Node, isExpanded bool, index int) *VisibleRow {
	return &VisibleRow{
		Node:         node,
		Depth:        node.Depth,
		IsExpanded:   isExpanded,
		IsExpandable: node.IsExpandable(),
		HasChildren:  node.HasChildren(),
		ChildCount:   node.ChildCount(),
		Index:        index,
	}
}

// DisplayKey returns the display key for this row
func (vr *VisibleRow) DisplayKey() string {
	return vr.Node.DisplayKey()
}

// Kind returns the node kind
func (vr *VisibleRow) Kind() NodeKind {
	return vr.Node.Kind
}

// ScalarValue returns the scalar value if this is a scalar node
func (vr *VisibleRow) ScalarValue() string {
	return vr.Node.ScalarValue
}

// ScalarType returns the scalar type if this is a scalar node
func (vr *VisibleRow) ScalarType() ScalarType {
	return vr.Node.ScalarType
}

// Path returns the path to this node
func (vr *VisibleRow) Path() *Path {
	return vr.Node.Path
}

// PathString returns the path as a string
func (vr *VisibleRow) PathString() string {
	if vr.Node.Path == nil {
		return "(root)"
	}
	return vr.Node.Path.String()
}
