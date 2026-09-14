package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func TestHTMLRendersDokuWikiLineBreak(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "erste Zeile\\\\ zweite Zeile"}}}
	got := HTML(p)
	if !strings.Contains(got, "erste Zeile<br/>zweite Zeile") {
		t.Fatalf("line break missing: %s", got)
	}
}

func TestPreviewHTMLRendersDokuWikiLineBreak(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "erste Zeile\\\\ zweite Zeile"}}}
	got := PreviewHTML(p)
	if !strings.Contains(got, "erste Zeile<br/>zweite Zeile") {
		t.Fatalf("line break missing: %s", got)
	}
}

func TestMarkdownRendersDokuWikiLineBreak(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "erste Zeile\\\\ zweite Zeile"}}}
	got := Markdown(p)
	if !strings.Contains(got, "erste Zeile<br/>zweite Zeile") {
		t.Fatalf("line break missing: %s", got)
	}
}

func TestDokuWikiLineBreakRequiresWhitespaceOrEndOfLine(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: "erste Zeile\\\\zweite Zeile"}}}
	for name, got := range map[string]string{
		"html": HTML(p), "preview": PreviewHTML(p), "markdown": Markdown(p),
	} {
		if strings.Contains(got, "erste Zeile<br/>zweite Zeile") {
			t.Fatalf("%s rendered an invalid line break: %s", name, got)
		}
	}
}
