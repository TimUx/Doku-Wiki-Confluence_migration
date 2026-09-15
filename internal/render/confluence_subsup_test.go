package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestConfluenceStorageRendersSingleCharacterSubscriptAndSuperscript(t *testing.T) {
	p := parser.Parse("demo:chemistry", "H_2O and x^2")
	got := ConfluenceStorage(p)
	if !strings.Contains(got, "H<sub>2</sub>O") { t.Fatalf("subscript missing: %s", got) }
	if !strings.Contains(got, "x<sup>2</sup>") { t.Fatalf("superscript missing: %s", got) }
}
func TestConfluenceStorageRendersBracedSubscriptAndSuperscript(t *testing.T) {
	p := parser.Parse("demo:math", "x_{foo} and y^{bar}")
	got := ConfluenceStorage(p)
	if !strings.Contains(got, "x<sub>{foo}</sub>") { t.Fatalf("braced subscript missing: %s", got) }
	if !strings.Contains(got, "y<sup>{bar}</sup>") { t.Fatalf("braced superscript missing: %s", got) }
}
func TestConfluenceStorageDoesNotFormatSubscriptInsideAttachmentFilename(t *testing.T) {
	p := parser.Parse("cc33:storage:howto:block:pure", "{{pure_architektur-uebericht.png}} {{pure_r4_rearview.png}}")
	got := ConfluenceStorage(p)
	if !strings.Contains(got, `ri:filename="cc33_storage_howto_block_pure_architektur-uebericht.png"`) { t.Fatalf("attachment filename was modified by subscript formatting: %s", got) }
	if !strings.Contains(got, `ri:filename="cc33_storage_howto_block_pure_r4_rearview.png"`) { t.Fatalf("second attachment filename was modified by subscript formatting: %s", got) }
	if strings.Contains(got, `ri:filename="cc33_storage_howto_block_pure_<sub>`) { t.Fatalf("subscript markup leaked into attachment filename: %s", got) }
}
