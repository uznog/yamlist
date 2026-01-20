package yamlparse

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/uznog/yamlist/internal/model"
	"gopkg.in/yaml.v3"
)

// ParseFile parses a YAML file and returns a Document
func ParseFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseBytes(data, path)
}

// ParseBytes parses YAML data from bytes
func ParseBytes(data []byte, sourcePath string) (*Document, error) {
	var yamlNode yaml.Node
	if err := yaml.Unmarshal(data, &yamlNode); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// yaml.Unmarshal wraps the content in a document node
	if yamlNode.Kind != yaml.DocumentNode || len(yamlNode.Content) == 0 {
		// Empty or invalid document - create empty root
		root := &model.Node{
			Kind:  model.KindMap,
			Path:  model.NewPath(),
			Depth: 0,
		}
		return NewDocument(root, sourcePath), nil
	}

	root := convertNode(yamlNode.Content[0], "", -1, 0, model.NewPath(), nil)
	return NewDocument(root, sourcePath), nil
}

// ParseString parses YAML from a string
func ParseString(data string) (*Document, error) {
	return ParseBytes([]byte(data), "<string>")
}

// convertNode recursively converts a yaml.Node to our internal Node
func convertNode(yn *yaml.Node, key string, index int, depth int, parentPath *model.Path, parent *model.Node) *model.Node {
	node := &model.Node{
		Key:        key,
		Index:      index,
		Depth:      depth,
		Parent:     parent,
		LineNumber: yn.Line,
	}

	// Build the path
	if key != "" {
		node.Path = parentPath.AppendKey(key)
	} else if index >= 0 {
		node.Path = parentPath.AppendIndex(index)
	} else {
		node.Path = parentPath
	}

	switch yn.Kind {
	case yaml.ScalarNode:
		node.Kind = model.KindScalar
		node.ScalarValue = yn.Value
		node.ScalarType = inferScalarType(yn)

	case yaml.MappingNode:
		node.Kind = model.KindMap
		node.Children = make([]*model.Node, 0, len(yn.Content)/2)

		// Mapping nodes have alternating key/value pairs
		for i := 0; i < len(yn.Content); i += 2 {
			keyNode := yn.Content[i]
			valueNode := yn.Content[i+1]
			child := convertNode(valueNode, keyNode.Value, -1, depth+1, node.Path, node)
			child.LineNumber = keyNode.Line // Use key's line, not value's line
			node.Children = append(node.Children, child)
		}

	case yaml.SequenceNode:
		node.Kind = model.KindList
		node.Children = make([]*model.Node, 0, len(yn.Content))

		for i, item := range yn.Content {
			child := convertNode(item, "", i, depth+1, node.Path, node)
			node.Children = append(node.Children, child)
		}

	case yaml.AliasNode:
		// Resolve aliases by following the reference
		if yn.Alias != nil {
			return convertNode(yn.Alias, key, index, depth, parentPath, parent)
		}
		// Fallback for unresolved alias
		node.Kind = model.KindScalar
		node.ScalarValue = "(alias)"
		node.ScalarType = model.ScalarString
	}

	return node
}

// inferScalarType infers the scalar type from the yaml.Node tag
func inferScalarType(yn *yaml.Node) model.ScalarType {
	// Check explicit tag first
	switch yn.Tag {
	case "!!null":
		return model.ScalarNull
	case "!!bool":
		return model.ScalarBool
	case "!!int":
		return model.ScalarInt
	case "!!float":
		return model.ScalarFloat
	case "!!timestamp":
		return model.ScalarTimestamp
	case "!!str":
		return model.ScalarString
	}

	// Infer from value if tag is not explicit
	value := yn.Value

	// Check for null
	if value == "" || value == "null" || value == "~" || value == "Null" || value == "NULL" {
		return model.ScalarNull
	}

	// Check for bool (YAML 1.2 only uses true/false)
	lowerVal := strings.ToLower(value)
	if lowerVal == "true" || lowerVal == "false" {
		return model.ScalarBool
	}

	// Check for int
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return model.ScalarInt
	}

	// Check for hex int
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		if _, err := strconv.ParseInt(value[2:], 16, 64); err == nil {
			return model.ScalarInt
		}
	}

	// Check for octal int
	if strings.HasPrefix(value, "0o") || strings.HasPrefix(value, "0O") {
		if _, err := strconv.ParseInt(value[2:], 8, 64); err == nil {
			return model.ScalarInt
		}
	}

	// Check for float
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		// Make sure it's actually a float (has decimal or exponent)
		if strings.Contains(value, ".") || strings.ContainsAny(value, "eE") {
			return model.ScalarFloat
		}
	}

	// Check for special float values
	if lowerVal == ".inf" || lowerVal == "-.inf" || lowerVal == ".nan" {
		return model.ScalarFloat
	}

	// Default to string
	return model.ScalarString
}
