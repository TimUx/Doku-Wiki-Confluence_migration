package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestConfluenceStorageRendersSingleCharacterSubscriptAndSuperscript(t *testing.T) {
	p := parser.Parse("demo:chemistry", "H_2O and x^2")
	got := ConfluenceStorage(p)
	if !strings.Contains(got, "H<sub>2</sub>O") {
		t.Fatalf("subscript missing: %s", got)
	}
	if !strings.Contains(got, "x<sup>2</sup>") {
		t.Fatalf("superscript missing: %s", got)
	}
}

func TestConfluenceStorageRendersBracedSubscriptAndSuperscript(t *testing.T) {
	p := parser.Parse("demo:math", "x_{foo} and y^{bar}")
	got := ConfluenceStorage(p)
	if !strings.Contains(got, "x<sub>{foo}</sub>") {
		t.Fatalf("braced subscript missing: %s", got)
	}
	if !strings.Contains(got, "y<sup>{bar}</sup>") {
		t.Fatalf("braced superscript missing: %s", got)
	}
}
