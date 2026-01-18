package model

// NodeKind represents the type of a YAML node
type NodeKind int

const (
	KindScalar NodeKind = iota
	KindMap
	KindList
)

func (k NodeKind) String() string {
	switch k {
	case KindScalar:
		return "scalar"
	case KindMap:
		return "map"
	case KindList:
		return "list"
	default:
		return "unknown"
	}
}

// ScalarType represents the inferred type of a scalar value
type ScalarType int

const (
	ScalarString ScalarType = iota
	ScalarInt
	ScalarFloat
	ScalarBool
	ScalarNull
	ScalarTimestamp
)

func (t ScalarType) String() string {
	switch t {
	case ScalarString:
		return "string"
	case ScalarInt:
		return "int"
	case ScalarFloat:
		return "float"
	case ScalarBool:
		return "bool"
	case ScalarNull:
		return "null"
	case ScalarTimestamp:
		return "timestamp"
	default:
		return "unknown"
	}
}

// Node represents a node in the YAML tree
type Node struct {
	// Key is the key name for this node (empty for root and list items)
	Key string

	// Kind indicates if this is a scalar, map, or list
	Kind NodeKind

	// ScalarValue holds the string value for scalar nodes
	ScalarValue string

	// ScalarType is the inferred type for scalar nodes
	ScalarType ScalarType

	// Children holds child nodes for maps and lists
	Children []*Node

	// Index is the list index for list items (-1 for non-list items)
	Index int

	// Depth is the nesting level (0 for root)
	Depth int

	// Path is the full path to this node
	Path *Path

	// Parent points to the parent node (nil for root)
	Parent *Node

	// LineNumber is the source line in the YAML file
	LineNumber int
}

// IsExpandable returns true if the node can have children
func (n *Node) IsExpandable() bool {
	return n.Kind == KindMap || n.Kind == KindList
}

// HasChildren returns true if the node has at least one child
func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

// ChildCount returns the number of children
func (n *Node) ChildCount() int {
	return len(n.Children)
}

// DisplayKey returns the display name for this node
func (n *Node) DisplayKey() string {
	if n.Key != "" {
		return n.Key
	}
	if n.Index >= 0 {
		return "[" + string(rune('0'+n.Index%10)) + "]"
	}
	return "(root)"
}
