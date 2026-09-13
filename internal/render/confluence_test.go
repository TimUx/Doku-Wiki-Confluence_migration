package render

import (
    "strings"
    "testing"

    "github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func TestConfluenceStorageKeepsIncludesAndUsesAttachments(t *testing.T) {
    p := model.Page{
        ID: "server:backup",
        Nodes: []model.Node{
            {Type: "heading", Level: "2", Text: "Backup"},
            {Type: "paragraph", Text: "Bild: {{images:backup.png|Backup}}"},
            {Type: "include", Target: "server:restore"},
        },
        Media: []model.Reference{{Target: "images:backup.png", Display: "Backup", Kind: "media"}},
        Includes: []model.Reference{{Target: "server:restore", Kind: "include"}},
    }
    got := ConfluenceStorage(p)
    if !strings.Contains(got, `ri:attachment ri:filename="backup.png"`) { t.Fatalf("attachment reference missing: %s", got) }
    if !strings.Contains(got, "[DOKUWIKI INCLUDE: server:restore]") { t.Fatalf("include placeholder missing: %s", got) }
}

func TestAttachmentNamesCollisionIsDeterministic(t *testing.T) {
    p := model.Page{Media: []model.Reference{{Target: "one:image.png"}, {Target: "two:image.png"}}}
    got := AttachmentNames(p)
    if got["one:image.png"] != "image.png" || got["two:image.png"] != "image_2.png" { t.Fatalf("unexpected names: %#v", got) }
}
