package model

import (
	"strconv"
	"strings"
)

// PathSegment represents a single segment in a path
type PathSegment struct {
	// Key is the map key (empty for list indices)
	Key string

	// Index is the list index (-1 for map keys)
	Index int
}

// IsIndex returns true if this segment represents a list index
func (s PathSegment) IsIndex() bool {
	return s.Index >= 0
}

// String returns the string representation of the segment
func (s PathSegment) String() string {
	if s.IsIndex() {
		return "[" + strconv.Itoa(s.Index) + "]"
	}
	return s.Key
}

// Path represents the full path to a node in the YAML tree
type Path struct {
	Segments []PathSegment
}

// NewPath creates a new empty path
func NewPath() *Path {
	return &Path{
		Segments: make([]PathSegment, 0),
	}
}

// Append creates a new path with the given segment appended
func (p *Path) Append(seg PathSegment) *Path {
	newPath := &Path{
		Segments: make([]PathSegment, len(p.Segments)+1),
	}
	copy(newPath.Segments, p.Segments)
	newPath.Segments[len(p.Segments)] = seg
	return newPath
}

// AppendKey creates a new path with the given key appended
func (p *Path) AppendKey(key string) *Path {
	return p.Append(PathSegment{Key: key, Index: -1})
}

// AppendIndex creates a new path with the given index appended
func (p *Path) AppendIndex(index int) *Path {
	return p.Append(PathSegment{Index: index})
}

// String returns the dot-notation string representation
// Example: "metadata.labels[0].name"
func (p *Path) String() string {
	if len(p.Segments) == 0 {
		return "(root)"
	}

	var b strings.Builder
	for i, seg := range p.Segments {
		if seg.IsIndex() {
			b.WriteString("[")
			b.WriteString(strconv.Itoa(seg.Index))
			b.WriteString("]")
		} else {
			if i > 0 && !p.Segments[i-1].IsIndex() {
				b.WriteString(".")
			} else if i > 0 {
				b.WriteString(".")
			}
			b.WriteString(seg.Key)
		}
	}
	return b.String()
}

// DisplayString returns a human-readable display string
// This is used for search matching
func (p *Path) DisplayString() string {
	return p.String()
}

// Depth returns the number of segments in the path
func (p *Path) Depth() int {
	return len(p.Segments)
}

// Parent returns a new path without the last segment
func (p *Path) Parent() *Path {
	if len(p.Segments) == 0 {
		return NewPath()
	}
	newPath := &Path{
		Segments: make([]PathSegment, len(p.Segments)-1),
	}
	copy(newPath.Segments, p.Segments[:len(p.Segments)-1])
	return newPath
}

// Equal returns true if two paths are equal
func (p *Path) Equal(other *Path) bool {
	if other == nil {
		return p == nil
	}
	if len(p.Segments) != len(other.Segments) {
		return false
	}
	for i, seg := range p.Segments {
		otherSeg := other.Segments[i]
		if seg.Key != otherSeg.Key || seg.Index != otherSeg.Index {
			return false
		}
	}
	return true
}

// IsAncestorOf returns true if this path is an ancestor of other
func (p *Path) IsAncestorOf(other *Path) bool {
	if other == nil || len(p.Segments) >= len(other.Segments) {
		return false
	}
	for i, seg := range p.Segments {
		otherSeg := other.Segments[i]
		if seg.Key != otherSeg.Key || seg.Index != otherSeg.Index {
			return false
		}
	}
	return true
}
