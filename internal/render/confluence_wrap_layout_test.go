package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func TestConfluenceStorageRendersWrapGroupColumnsWidthsAndAlignment(t *testing.T) {
	p := model.Page{ID: "layout", Nodes: []model.Node{{
		Type:   "wrap",
		Plugin: "wrap",
		Meta:   map[string]string{"kind": "group", "classes": "group"},
		Children: []model.Node{
			{Type: "wrap", Plugin: "wrap", Meta: map[string]string{"kind": "column", "classes": "column left", "width": "25%"}, Children: []model.Node{{Type: "paragraph", Text: "Left"}}},
			{Type: "wrap", Plugin: "wrap", Meta: map[string]string{"kind": "column", "classes": "column center", "width": "50%"}, Children: []model.Node{{Type: "paragraph", Text: "Center"}}},
			{Type: "wrap", Plugin: "wrap", Meta: map[string]string{"kind": "column", "classes": "column", "width": "25%", "align": "right"}, Children: []model.Node{{Type: "paragraph", Text: "Right"}}},
		},
	}}}

	got := ConfluenceStorage(p)
	if strings.Count(got, `ac:name="column"`) != 3 {
		t.Fatalf("expected three native column macros: %s", got)
	}
	for _, width := range []string{"25%", "50%", "25%"} {
		if !strings.Contains(got, `<ac:parameter ac:name="width">`+width+`</ac:parameter>`) {
			t.Fatalf("column width %s missing: %s", width, got)
		}
	}
	for _, align := range []string{"text-align:left;", "text-align:center;", "text-align:right;"} {
		if !strings.Contains(got, `style="`+align+`"`) {
			t.Fatalf("column alignment %s missing: %s", align, got)
		}
	}
	if !strings.Contains(got, `<ac:structured-macro ac:name="section">`) {
		t.Fatalf("native section macro missing: %s", got)
	}
	if !strings.Contains(got, `<p>Left</p>`) || !strings.Contains(got, `<p>Center</p>`) || !strings.Contains(got, `<p>Right</p>`) {
		t.Fatalf("column bodies missing: %s", got)
	}
}

func TestConfluenceStorageRendersWrapLayoutSyntaxAndClear(t *testing.T) {
	src := `<WRAP group>
<WRAP column 50% left>
Left **content**
</WRAP>
<WRAP column width=50% align=right>
Right
</WRAP>
</WRAP>
<WRAP clear />`
	p := parseForWrapLayoutTest(src)
	got := ConfluenceStorage(p)

	if !strings.Contains(got, `<ac:structured-macro ac:name="section">`) {
		t.Fatalf("group did not map to section: %s", got)
	}
	if strings.Count(got, `ac:name="column"`) != 2 {
		t.Fatalf("expected two columns: %s", got)
	}
	if !strings.Contains(got, `<ac:parameter ac:name="width">50%</ac:parameter>`) {
		t.Fatalf("50%% width missing: %s", got)
	}
	if !strings.Contains(got, `style="text-align:left;"`) || !strings.Contains(got, `style="text-align:right;"`) {
		t.Fatalf("left/right alignment missing: %s", got)
	}
	if !strings.Contains(got, `<strong>content</strong>`) {
		t.Fatalf("inline content inside column was not rendered: %s", got)
	}
	if !strings.Contains(got, `<div style="clear:both;"></div>`) {
		t.Fatalf("clear layout marker missing: %s", got)
	}
}

func parseForWrapLayoutTest(src string) model.Page {
	return parseForWrapLayoutSource(src)
}
