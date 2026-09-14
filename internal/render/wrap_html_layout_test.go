package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestHTMLRendersWrapGroupColumnsWidthsAndAlignment(t *testing.T) {
	src := `<WRAP group>
<WRAP column 25% left>
Left
</WRAP>
<WRAP column 50% center>
Center
</WRAP>
<WRAP column width=25% align=right>
Right
</WRAP>
</WRAP>`
	p := parser.Parse("layout", src)
	got := HTML(p)

	if !strings.Contains(got, `class="dokuwiki-wrap group" style="display:flex;flex-wrap:wrap;width:100%;"`) {
		t.Fatalf("group flex layout missing: %s", got)
	}
	for _, want := range []string{
		`class="dokuwiki-wrap column left" style="width:25%;text-align:left;"`,
		`class="dokuwiki-wrap column center" style="width:50%;text-align:center;"`,
		`class="dokuwiki-wrap column" style="width:25%;text-align:right;"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("column layout missing %q: %s", want, got)
		}
	}
}

func TestPreviewHTMLRendersNestedWrapLayoutAndClear(t *testing.T) {
	src := `<WRAP group>
<WRAP column 50% left>
**Left**
<WRAP group>
<WRAP column 100% center>Nested</WRAP>
</WRAP>
</WRAP>
<WRAP column 50% right>Right</WRAP>
</WRAP>
<WRAP clear />`
	p := parser.Parse("layout", src)
	got := PreviewHTML(p)

	if strings.Count(got, `class="dokuwiki-wrap group"`) != 2 {
		t.Fatalf("expected nested group wrappers: %s", got)
	}
	if !strings.Contains(got, `<strong>Left</strong>`) || !strings.Contains(got, `Nested`) || !strings.Contains(got, `Right`) {
		t.Fatalf("nested wrap content missing: %s", got)
	}
	if !strings.Contains(got, `class="dokuwiki-wrap column right" style="width:50%;text-align:right;"`) {
		t.Fatalf("right aligned column missing: %s", got)
	}
	if !strings.Contains(got, `<div class="dokuwiki-clear" style="clear:both;"></div>`) {
		t.Fatalf("clear marker missing: %s", got)
	}
}

func TestHTMLIgnoresUnsafeWrapWidth(t *testing.T) {
	src := `<WRAP column 50%;color:red left>Unsafe</WRAP>`
	p := parser.Parse("layout", src)
	got := HTML(p)
	if strings.Contains(got, `style="width:50%;color:red`) || strings.Contains(got, `style="color:red`) {
		t.Fatalf("unsafe width leaked into HTML style: %s", got)
	}
	if !strings.Contains(got, `text-align:left;`) {
		t.Fatalf("safe alignment should still render: %s", got)
	}
}
