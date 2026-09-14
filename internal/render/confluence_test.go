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
    if !strings.Contains(got, `<ac:structured-macro ac:name="toc"/>`) { t.Fatalf("toc macro missing: %s", got) }
}

func TestAttachmentNamesCollisionIsDeterministic(t *testing.T) {
    p := model.Page{Media: []model.Reference{{Target: "one:image.png"}, {Target: "two:image.png"}}}
    got := AttachmentNames(p)
    if got["one:image.png"] != "image.png" || got["two:image.png"] != "image_2.png" { t.Fatalf("unexpected names: %#v", got) }
}

func TestConfluenceStorageRendersExternalAndSamePageLinks(t *testing.T) {
    p := model.Page{
        ID: "storage:security",
        Nodes: []model.Node{
            {Type: "heading", Level: "2", Text: "Datensicherheit"},
            {Type: "paragraph", Text: `[[https://example.com/security.pdf|Security PDF]] [[#datensicherheit|Zum Abschnitt]] [[storage:other|Andere Seite]]`},
        },
    }
    got := ConfluenceStorage(p)
    if !strings.Contains(got, `<a href="https://example.com/security.pdf">Security PDF</a>`) { t.Fatalf("external link missing: %s", got) }
    if !strings.Contains(got, `<a href="#datensicherheit">Zum Abschnitt</a>`) { t.Fatalf("same-page link missing: %s", got) }
    if !strings.Contains(got, `[DOKUWIKI LINK: storage:other]`) { t.Fatalf("internal-page placeholder missing: %s", got) }
}

func TestPreviewRendersNestedMediaLink(t *testing.T) {
    p := model.Page{
        ID: "cc33:storage:howto:block:pure:active-cluster",
        Nodes: []model.Node{{Type: "paragraph", Text: `[[https://example.com/wiki/lib/exe/detail.php/cc33:storage:howto:block:pure:pod_1.png|{{https://example.com/wiki/lib/exe/fetch.php/cc33:storage:howto:block:pure:pod_1.png?400}}]]`}},
    }
    got := PreviewHTML(p)
    if strings.Contains(got, `[[https://`) { t.Fatalf("raw wiki link remains: %s", got) }
    if !strings.Contains(got, `<a href="https://example.com/wiki/lib/exe/detail.php/cc33:storage:howto:block:pure:pod_1.png"`) { t.Fatalf("image detail link missing: %s", got) }
    if !strings.Contains(got, `/api/media?target=cc33%3Astorage%3Ahowto%3Ablock%3Apure%3Apod_1.png`) { t.Fatalf("media endpoint target missing: %s", got) }
}
