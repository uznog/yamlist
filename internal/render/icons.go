package render

import "github.com/vznog/yamlist/internal/model"

// IconSet defines the icons used for rendering
type IconSet struct {
	// Expand/collapse markers
	Expanded   string
	Collapsed  string
	Leaf       string

	// Node type icons
	Map        string
	List       string
	String     string
	Number     string
	Bool       string
	Null       string
	Timestamp  string

	// Tree lines
	Connector  string
	LastItem   string
	Vertical   string
	Empty      string
}

// NerdFontIcons returns icons using Nerd Font symbols
func NerdFontIcons() *IconSet {
	return &IconSet{
		Expanded:   "▾",
		Collapsed:  "▸",
		Leaf:       " ",

		Map:        "",
		List:       "",
		String:     "",
		Number:     "󰎠",
		Bool:       "󰨙",
		Null:       "󰟢",
		Timestamp:  "",

		Connector:  "├",
		LastItem:   "└",
		Vertical:   "│",
		Empty:      " ",
	}
}

// ASCIIIcons returns ASCII-only icons as fallback
func ASCIIIcons() *IconSet {
	return &IconSet{
		Expanded:   "v",
		Collapsed:  ">",
		Leaf:       " ",

		Map:        "{}",
		List:       "[]",
		String:     "\"",
		Number:     "#",
		Bool:       "?",
		Null:       "~",
		Timestamp:  "@",

		Connector:  "|-",
		LastItem:   "`-",
		Vertical:   "| ",
		Empty:      "  ",
	}
}

// GetExpandIcon returns the appropriate expand/collapse icon
func (icons *IconSet) GetExpandIcon(isExpanded bool, isExpandable bool) string {
	if !isExpandable {
		return icons.Leaf
	}
	if isExpanded {
		return icons.Expanded
	}
	return icons.Collapsed
}

// GetTypeIcon returns the icon for a node type
func (icons *IconSet) GetTypeIcon(kind model.NodeKind, scalarType model.ScalarType) string {
	switch kind {
	case model.KindMap:
		return icons.Map
	case model.KindList:
		return icons.List
	case model.KindScalar:
		return icons.GetScalarIcon(scalarType)
	default:
		return " "
	}
}

// GetScalarIcon returns the icon for a scalar type
func (icons *IconSet) GetScalarIcon(scalarType model.ScalarType) string {
	switch scalarType {
	case model.ScalarString:
		return icons.String
	case model.ScalarInt, model.ScalarFloat:
		return icons.Number
	case model.ScalarBool:
		return icons.Bool
	case model.ScalarNull:
		return icons.Null
	case model.ScalarTimestamp:
		return icons.Timestamp
	default:
		return icons.String
	}
}
