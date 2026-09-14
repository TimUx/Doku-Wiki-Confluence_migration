package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func TestHTMLRendersSubscriptAndSuperscript(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "H_2O and x^2"}}}
	got := HTML(p)
	if !strings.Contains(got, "H<sub>2</sub>O") {
		t.Fatalf("subscript missing: %s", got)
	}
	if !strings.Contains(got, "x<sup>2</sup>") {
		t.Fatalf("superscript missing: %s", got)
	}
}

func TestPreviewHTMLRendersSubscriptAndSuperscript(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "H_2O and x^2"}}}
	got := PreviewHTML(p)
	if !strings.Contains(got, "H<sub>2</sub>O") {
		t.Fatalf("subscript missing: %s", got)
	}
	if !strings.Contains(got, "x<sup>2</sup>") {
		t.Fatalf("superscript missing: %s", got)
	}
}

func TestMarkdownRendersSubscriptAndSuperscript(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "H_2O and x^2"}}}
	got := Markdown(p)
	if !strings.Contains(got, "H<sub>2</sub>O") {
		t.Fatalf("subscript missing: %s", got)
	}
	if !strings.Contains(got, "x<sup>2</sup>") {
		t.Fatalf("superscript missing: %s", got)
	}
}
