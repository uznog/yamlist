package yamlparse

import "github.com/uznog/yamlist/internal/model"

// Document represents a parsed YAML document
type Document struct {
	// Root is the root node of the YAML tree
	Root *model.Node

	// Index is the flattened path index for searching
	Index *model.PathIndex

	// FilePath is the path to the source file
	FilePath string
}

// NewDocument creates a new document with the given root
func NewDocument(root *model.Node, filePath string) *Document {
	doc := &Document{
		Root:     root,
		Index:    model.NewPathIndex(),
		FilePath: filePath,
	}
	doc.buildIndex(root)
	return doc
}

// buildIndex recursively builds the path index
func (d *Document) buildIndex(node *model.Node) {
	if node == nil {
		return
	}

	// Add this node to the index
	entry := model.NewPathEntry(node)
	d.Index.Add(entry)

	// Recursively process children
	for _, child := range node.Children {
		d.buildIndex(child)
	}
}

// NodeCount returns the total number of nodes in the document
func (d *Document) NodeCount() int {
	return d.Index.Len()
}

// FindByPath finds a node by its path string
func (d *Document) FindByPath(pathStr string) *model.Node {
	for _, entry := range d.Index.Entries() {
		if entry.DisplayString == pathStr {
			return entry.Node
		}
	}
	return nil
}
