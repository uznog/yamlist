package render

import (
	"regexp"
	"strings"
	"testing"

	"github.com/vznog/yamlist/internal/model"
)

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansi.ReplaceAllString(s, "")
}

func TestFormatScalarValue_Multiline(t *testing.T) {
	r := NewRowRenderer(ASCIIIcons(), DefaultStyles())

	tests := []struct {
		name     string
		value    string
		expected string // substring that must be present
	}{
		{"folded", "Line 1\nLine 2\nLine 3", "[3 lines]"},
		{"literal", "#!/bin/bash\nset -e\necho hello", "[3 lines]"},
		{"escaped_newlines", "single\\nline", "single\\nline"}, // not actual newline
		{"no_newlines", "simple value", "simple value"},
		{"single_line_with_trailing_newline", "just one\n", "[2 lines]"},
		{"many_lines", "a\nb\nc\nd\ne\nf", "[6 lines]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.formatScalarValue(tt.value, model.ScalarString, false)
			// Strip ANSI codes for comparison
			plain := stripANSI(result)
			if !strings.Contains(plain, tt.expected) {
				t.Errorf("expected %q to contain %q", plain, tt.expected)
			}
			// Verify no actual newlines in output
			if strings.Contains(plain, "\n") {
				t.Errorf("output contains literal newline: %q", plain)
			}
		})
	}
}

func TestFormatScalarValue_Truncation(t *testing.T) {
	r := NewRowRenderer(ASCIIIcons(), DefaultStyles())

	tests := []struct {
		name     string
		value    string
		contains string
	}{
		{"long_string_truncated", strings.Repeat("a", 100), "..."},
		{"short_string_not_truncated", "short", "short"},
		{"unicode_truncation", strings.Repeat("日本語", 20), "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.formatScalarValue(tt.value, model.ScalarString, false)
			plain := stripANSI(result)
			if !strings.Contains(plain, tt.contains) {
				t.Errorf("expected %q to contain %q", plain, tt.contains)
			}
		})
	}
}

func TestFormatScalarValue_Types(t *testing.T) {
	r := NewRowRenderer(ASCIIIcons(), DefaultStyles())

	tests := []struct {
		name       string
		value      string
		scalarType model.ScalarType
		expected   string
	}{
		{"null", "", model.ScalarNull, "null"},
		{"bool_true", "true", model.ScalarBool, "true"},
		{"int", "42", model.ScalarInt, "42"},
		{"float", "3.14", model.ScalarFloat, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.formatScalarValue(tt.value, tt.scalarType, false)
			plain := stripANSI(result)
			if !strings.Contains(plain, tt.expected) {
				t.Errorf("expected %q to contain %q", plain, tt.expected)
			}
		})
	}
}

func TestRuneCount(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"日本語", 3},
		{"hello日本語", 8},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := runeCount(tt.input)
			if got != tt.expected {
				t.Errorf("runeCount(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		input    string
		maxRunes int
		expected string
	}{
		{"hello", 3, "hel"},
		{"hello", 10, "hello"},
		{"日本語", 2, "日本"},
		{"hello日本語", 7, "hello日本"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncateRunes(tt.input, tt.maxRunes)
			if got != tt.expected {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.maxRunes, got, tt.expected)
			}
		})
	}
}
