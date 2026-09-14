package render

import (
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestMarkdownPreservesCodeLanguage(t *testing.T) {
	p := parser.Parse("demo:code", "<code go>\nfmt.Println(\"hello\")\n</code>")
	got := Markdown(p)
	want := "```go\nfmt.Println(\"hello\")\n```"
	if !strings.Contains(got, want) {
		t.Fatalf("code language missing: %s", got)
	}
}

func TestMarkdownPreservesFileLanguage(t *testing.T) {
	p := parser.Parse("demo:file", "<file yaml>\nservices:\n  app:\n    image: nginx\n</file>")
	got := Markdown(p)
	want := "```yaml\nservices:\n  app:\n    image: nginx\n```"
	if !strings.Contains(got, want) {
		t.Fatalf("file language missing: %s", got)
	}
}
