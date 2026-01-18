package yamlparse

import (
	"testing"

	"github.com/uznog/yamlist/internal/model"
)

func TestParseString_Simple(t *testing.T) {
	yaml := `
name: John
age: 30
active: true
`
	doc, err := ParseString(yaml)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if doc.Root == nil {
		t.Fatal("Root is nil")
	}

	if doc.Root.Kind != model.KindMap {
		t.Errorf("Expected root to be Map, got %v", doc.Root.Kind)
	}

	if len(doc.Root.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(doc.Root.Children))
	}

	// Check first child
	name := doc.Root.Children[0]
	if name.Key != "name" {
		t.Errorf("Expected key 'name', got '%s'", name.Key)
	}
	if name.ScalarValue != "John" {
		t.Errorf("Expected value 'John', got '%s'", name.ScalarValue)
	}
	if name.ScalarType != model.ScalarString {
		t.Errorf("Expected ScalarString, got %v", name.ScalarType)
	}

	// Check second child
	age := doc.Root.Children[1]
	if age.Key != "age" {
		t.Errorf("Expected key 'age', got '%s'", age.Key)
	}
	if age.ScalarType != model.ScalarInt {
		t.Errorf("Expected ScalarInt, got %v", age.ScalarType)
	}

	// Check third child
	active := doc.Root.Children[2]
	if active.Key != "active" {
		t.Errorf("Expected key 'active', got '%s'", active.Key)
	}
	if active.ScalarType != model.ScalarBool {
		t.Errorf("Expected ScalarBool, got %v", active.ScalarType)
	}
}

func TestParseString_Nested(t *testing.T) {
	yaml := `
database:
  host: localhost
  port: 5432
`
	doc, err := ParseString(yaml)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if len(doc.Root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Root.Children))
	}

	db := doc.Root.Children[0]
	if db.Key != "database" {
		t.Errorf("Expected key 'database', got '%s'", db.Key)
	}
	if db.Kind != model.KindMap {
		t.Errorf("Expected Map, got %v", db.Kind)
	}
	if len(db.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(db.Children))
	}

	// Check path
	if db.Path.String() != "database" {
		t.Errorf("Expected path 'database', got '%s'", db.Path.String())
	}

	host := db.Children[0]
	if host.Path.String() != "database.host" {
		t.Errorf("Expected path 'database.host', got '%s'", host.Path.String())
	}
}

func TestParseString_List(t *testing.T) {
	yaml := `
items:
  - first
  - second
  - third
`
	doc, err := ParseString(yaml)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	items := doc.Root.Children[0]
	if items.Kind != model.KindList {
		t.Errorf("Expected List, got %v", items.Kind)
	}
	if len(items.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(items.Children))
	}

	// Check list item paths
	for i, child := range items.Children {
		if child.Index != i {
			t.Errorf("Expected index %d, got %d", i, child.Index)
		}
	}

	// Check path format
	first := items.Children[0]
	if first.Path.String() != "items[0]" {
		t.Errorf("Expected path 'items[0]', got '%s'", first.Path.String())
	}
}

func TestParseString_Index(t *testing.T) {
	yaml := `
name: test
value: 42
`
	doc, err := ParseString(yaml)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	// Check that index was built
	if doc.Index.Len() != 3 { // root + 2 children
		t.Errorf("Expected 3 entries in index, got %d", doc.Index.Len())
	}

	// Check display strings
	strings := doc.Index.DisplayStrings()
	found := make(map[string]bool)
	for _, s := range strings {
		found[s] = true
	}

	if !found["name"] {
		t.Error("Index missing 'name'")
	}
	if !found["value"] {
		t.Error("Index missing 'value'")
	}
}

func TestParseFile(t *testing.T) {
	doc, err := ParseFile("../../testdata/simple.yaml")
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if doc.Root == nil {
		t.Fatal("Root is nil")
	}

	if len(doc.Root.Children) != 5 {
		t.Errorf("Expected 5 children, got %d", len(doc.Root.Children))
	}
}

func TestInferScalarType(t *testing.T) {
	tests := []struct {
		yaml     string
		expected model.ScalarType
	}{
		{"value: hello", model.ScalarString},
		{"value: 42", model.ScalarInt},
		{"value: 3.14", model.ScalarFloat},
		{"value: true", model.ScalarBool},
		{"value: false", model.ScalarBool},
		{"value: null", model.ScalarNull},
		{"value: ~", model.ScalarNull},
		// Note: "yes"/"no" are NOT booleans in YAML 1.2 (yaml.v3)
		{"value: yes", model.ScalarString},
		{"value: no", model.ScalarString},
	}

	for _, tt := range tests {
		doc, err := ParseString(tt.yaml)
		if err != nil {
			t.Errorf("ParseString(%q) failed: %v", tt.yaml, err)
			continue
		}

		if len(doc.Root.Children) == 0 {
			t.Errorf("ParseString(%q): no children", tt.yaml)
			continue
		}

		got := doc.Root.Children[0].ScalarType
		if got != tt.expected {
			t.Errorf("ParseString(%q): expected %v, got %v", tt.yaml, tt.expected, got)
		}
	}
}
