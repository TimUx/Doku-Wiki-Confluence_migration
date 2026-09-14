package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

const noFormatInput = `%%**not bold** [[demo:page|not a link]] {{image.png}} H_2O%% and **bold**`

func TestHTMLPreservesDokuWikiNoFormat(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: noFormatInput}}}
	got := HTML(p)
	want := `**not bold** [[demo:page|not a link]] {{image.png}} H_2O`
	if !strings.Contains(got, want) {
		t.Fatalf("no-format content changed: %s", got)
	}
	if strings.Contains(got, "<strong>not bold</strong>") || strings.Contains(got, `href=`) || strings.Contains(got, `dokuwiki-media`) {
		t.Fatalf("no-format content was interpreted: %s", got)
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Fatalf("normal formatting stopped working: %s", got)
	}
}

func TestPreviewHTMLPreservesDokuWikiNoFormat(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: noFormatInput}}}
	got := PreviewHTML(p)
	want := `**not bold** [[demo:page|not a link]] {{image.png}} H_2O`
	if !strings.Contains(got, want) {
		t.Fatalf("no-format content changed: %s", got)
	}
	if strings.Contains(got, "<strong>not bold</strong>") || strings.Contains(got, `href=`) || strings.Contains(got, `dokuwiki-media`) {
		t.Fatalf("no-format content was interpreted: %s", got)
	}
}

func TestMarkdownPreservesDokuWikiNoFormat(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: noFormatInput}}}
	got := Markdown(p)
	if !strings.Contains(got, `\*\*not bold\*\*`) {
		t.Fatalf("markdown would interpret no-format bold markers: %s", got)
	}
	if !strings.Contains(got, `\[\[demo:page\|not a link\]\]`) {
		t.Fatalf("markdown changed no-format link syntax: %s", got)
	}
	if !strings.Contains(got, `{{image.png}} H\_2O`) {
		t.Fatalf("markdown changed no-format literal syntax: %s", got)
	}
	if !strings.Contains(got, "**bold**") {
		t.Fatalf("normal DokuWiki formatting was not preserved for markdown conversion: %s", got)
	}
}
