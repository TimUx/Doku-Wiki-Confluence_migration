package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestHTMLRendersContiguousTableAsSingleTable(t *testing.T) {
	p := parser.Parse("demo:table", `^ Name ^ Value ^
| Alpha | 1 |
| Beta | 2 |`)
	got := HTML(p)
	if strings.Count(got, `<table class="dokuwiki-table">`) != 1 {
		t.Fatalf("expected one table, got: %s", got)
	}
	if !strings.Contains(got, "<thead><tr><th>Name</th><th>Value</th></tr></thead>") {
		t.Fatalf("expected semantic header row, got: %s", got)
	}
	if !strings.Contains(got, "<tbody><tr><td>Alpha</td><td>1</td></tr><tr><td>Beta</td><td>2</td></tr></tbody>") {
		t.Fatalf("expected all body rows in one tbody, got: %s", got)
	}
}

func TestPreviewHTMLRendersTableCellLinksAndColspan(t *testing.T) {
	p := parser.Parse("demo:table", `^ Name ^ Value ^
| [[https://example.com]] | ::: |
| Combined | 2 |`)
	got := PreviewHTML(p)
	if !strings.Contains(got, `href="https://example.com"`) {
		t.Fatalf("expected preview link, got: %s", got)
	}
	if !strings.Contains(got, `colspan="2"`) {
		t.Fatalf("expected colspan, got: %s", got)
	}
	if strings.Count(got, `<table class="dokuwiki-table">`) != 1 {
		t.Fatalf("expected one preview table, got: %s", got)
	}
}

func TestMarkdownRendersHeaderSeparator(t *testing.T) {
	p := parser.Parse("demo:table", `^ Name ^ Value ^
| Alpha | one |
| Beta | 2 |`)
	got := Markdown(p)
	if !strings.Contains(got, "| Name | Value |\n| --- | --- |\n") {
		t.Fatalf("expected markdown header separator, got: %s", got)
	}
	if !strings.Contains(got, "| Alpha | one |\n| Beta | 2 |\n") {
		t.Fatalf("expected markdown body rows, got: %s", got)
	}
}
