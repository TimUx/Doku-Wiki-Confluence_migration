package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestConfluenceStoragePreservesMediaDimensions(t *testing.T) {
	p := parser.Parse("cc33:storage:howto:block:pure", `{{:cc33:storage:howto:block:pure:architektur.png?direct&800|}} {{:cc33:storage:howto:block:pure:detail.png?400x250|Detail}}`)
	got := ConfluenceStorage(p)
	if !strings.Contains(got, `<ac:image ac:width="800"><ri:attachment ri:filename="cc33_storage_howto_block_pure_architektur.png"/></ac:image>`) {
		t.Fatalf("single media width was not preserved: %s", got)
	}
	if !strings.Contains(got, `<ac:image ac:width="400" ac:height="250"><ri:attachment ri:filename="cc33_storage_howto_block_pure_detail.png"/></ac:image>`) {
		t.Fatalf("media width/height was not preserved: %s", got)
	}
}
