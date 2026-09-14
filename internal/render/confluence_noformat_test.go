package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func TestConfluenceStoragePreservesDokuWikiNoFormat(t *testing.T) {
	p := model.Page{ID: "demo", Nodes: []model.Node{{Type: "paragraph", Text: noFormatInput}}}
	got := ConfluenceStorage(p)
	want := `**not bold** [[demo:page|not a link]] {{image.png}} H_2O`
	if !strings.Contains(got, want) {
		t.Fatalf("no-format content changed: %s", got)
	}
	if strings.Contains(got, "<strong>not bold</strong>") || strings.Contains(got, "<ac:image>") || strings.Contains(got, "<span>[DOKUWIKI LINK:") {
		t.Fatalf("no-format content was interpreted: %s", got)
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Fatalf("normal formatting stopped working: %s", got)
	}
}
