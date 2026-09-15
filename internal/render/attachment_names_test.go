package render

import (
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
)

func TestAttachmentNamesUseResolvedNamespacePath(t *testing.T) {
	p := parser.Parse("server:backup", `{{:images:Architektur.png}} {{:docs:Architektur.png}}`)
	names := AttachmentNames(p)
	if names["images:Architektur.png"] != "images_Architektur.png" {
		t.Fatalf("unexpected images attachment name: %q", names["images:Architektur.png"])
	}
	if names["docs:Architektur.png"] != "docs_Architektur.png" {
		t.Fatalf("unexpected docs attachment name: %q", names["docs:Architektur.png"])
	}
	if names["images:Architektur.png"] == names["docs:Architektur.png"] {
		t.Fatal("attachments with the same basename must remain distinct")
	}
}

func TestAttachmentNamesKeepExactTargetStable(t *testing.T) {
	p := parser.Parse("server:backup", `{{:images:Architektur.png}} {{:images:Architektur.png}}`)
	names := AttachmentNames(p)
	if names["images:Architektur.png"] != "images_Architektur.png" {
		t.Fatalf("repeated target must keep one stable filename: %q", names["images:Architektur.png"])
	}
}
