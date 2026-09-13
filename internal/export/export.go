package export

import (
    "archive/zip"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
    "time"

    "github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
    "github.com/TimUx/Doku-Wiki-Confluence_migration/internal/render"
    "github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

var safe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type Manifest struct {
    ExportVersion int               `json:"exportVersion"`
    Created       string            `json:"created"`
    Source        map[string]string `json:"source"`
    Target        map[string]string `json:"target"`
    Pages         []model.Page      `json:"pages"`
}

func Create(db *store.Store, dir, name string, ids []string) (string, error) {
    if len(ids) == 0 { return "", fmt.Errorf("select at least one page") }
    name = safe.ReplaceAllString(name, "-"); if name == "" { name = "migration" }
    path := filepath.Join(dir, name+".zip")
    f, e := os.Create(path); if e != nil { return "", e }; defer f.Close()
    zw := zip.NewWriter(f)
    var pages []model.Page
    seen := map[string]bool{}
    for _, id := range ids {
        if seen[id] { continue }; seen[id] = true
        p, err := db.Page(id); if err != nil { return "", err }
        pages = append(pages, p)
    }
    sort.Slice(pages, func(i, j int) bool { return pages[i].ID < pages[j].ID })
    manifest := Manifest{1, time.Now().Format(time.RFC3339), map[string]string{"type":"dokuwiki"}, map[string]string{"type":"confluence", "version":"10.2.17"}, pages}
    b, _ := json.MarshalIndent(manifest, "", "  "); write(zw, "manifest.json", b)

    for i, p := range pages {
        base := fmt.Sprintf("pages/%03d_%s/", i+1, safe.ReplaceAllString(p.Title, "_"))
        names := render.AttachmentNames(p)
        write(zw, base+"page.html", []byte(render.HTML(p)))
        write(zw, base+"page.md", []byte(render.Markdown(p)))
        write(zw, base+"confluence-storage.xml", []byte(render.ConfluenceStorage(p)))
        write(zw, base+"source.txt", []byte(p.Source))
        meta, _ := json.MarshalIndent(p, "", "  "); write(zw, base+"metadata.json", meta)

        var attachments [][]string
        for _, m := range p.Media {
            src, err := db.MediaPath(m.Target); if err != nil { continue }
            in, err := os.Open(src); if err != nil { continue }
            filename := names[m.Target]; if filename == "" { filename = filepath.Base(src) }
            w, err := zw.Create(base+"attachments/"+filename)
            if err == nil { _, _ = io.Copy(w, in) }
            in.Close()
            attachments = append(attachments, []string{m.Target, filename, m.MIME})
        }
        writeCSV(zw, base+"attachments.csv", []string{"DokuWiki-Ziel", "Confluence-Anhangsname", "MIME"}, attachments)

        var includes [][]string
        for _, inc := range p.Includes { includes = append(includes, []string{p.ID, inc.Target, "[DOKUWIKI INCLUDE: "+inc.Target+"]"}) }
        writeCSV(zw, base+"includes.csv", []string{"Quellseite", "DokuWiki-Include", "Platzhalter"}, includes)
    }

    write(zw, "MIGRATION-GUIDE.md", []byte(guideMarkdown(pages)))
    write(zw, "MIGRATION-GUIDE.html", []byte(guideHTML(pages)))
    report := fmt.Sprintf("<!doctype html><meta charset=utf-8><title>Migration report</title><h1>Migration Summary</h1><p>Pages: %d</p><p>Includes bleiben als Platzhalter erhalten.</p><p>Bilder werden mit den exportierten Anhangsnamen referenziert.</p><p>Generated: %s</p>", len(pages), htmlEscape(time.Now().Format(time.RFC3339)))
    write(zw, "migration-report.html", []byte(report))
    rb, _ := json.MarshalIndent(map[string]any{"pages": len(pages), "generated": time.Now(), "includes":"placeholder", "attachments":"confluence-reference"}, "", "  ")
    write(zw, "migration-report.json", rb)
    if e = zw.Close(); e != nil { return "", e }
    db.RecordExport(path, len(pages)); return path, nil
}

func guideMarkdown(pages []model.Page) string {
    var b strings.Builder
    b.WriteString("# DokuWiki → Confluence 10.2 Migration Guide\n\n")
    b.WriteString("## Grundprinzip\n\nDie Exportdateien sind für eine kontrollierte manuelle Übernahme vorbereitet. Die zukünftige Confluence-Struktur wird nicht geraten. Includes bleiben deshalb als `[DOKUWIKI INCLUDE: ...]` markiert. Bilder verweisen bereits auf den exakten exportierten Anhangsnamen.\n\n")
    b.WriteString("## Ablauf\n\n1. Zielseite in Confluence anlegen und Titel prüfen.\n2. Alle Dateien aus dem jeweiligen `attachments/`-Ordner hochladen. Dateinamen **nicht ändern**.\n3. `confluence-storage.xml` als technische Referenz für die vorbereitete Confluence-Struktur verwenden.\n4. Inhalt in die Zielseite übernehmen und Formatierung prüfen.\n5. Jeden Include-Platzhalter bewusst durch die später passende Confluence-Referenz bzw. das Include Page Macro ersetzen.\n6. Interne Link-Platzhalter anhand der finalen Seitenstruktur auflösen.\n7. Bilder, Tabellen, Codeblöcke und Links in der Zielseite kontrollieren.\n\n")
    b.WriteString("## Exportierte Seiten\n\n")
    for _, p := range pages {
        fmt.Fprintf(&b, "- `%s` — %s — %d Medien, %d Includes, %d Warnungen\n", p.ID, p.Title, len(p.Media), len(p.Includes), len(p.Warnings))
    }
    b.WriteString("\n## Wichtig\n\nDer Export entscheidet nicht, wo eine Seite künftig in Confluence liegt. Diese Zuordnung erfolgt bewusst erst bei der Übernahme.\n")
    return b.String()
}

func guideHTML(pages []model.Page) string {
    var b strings.Builder
    b.WriteString("<!doctype html><meta charset=utf-8><title>DokuWiki → Confluence Migration Guide</title><h1>DokuWiki → Confluence 10.2</h1><h2>Übernahme</h2><ol><li>Zielseite anlegen.</li><li>Alle Dateien aus <code>attachments/</code> hochladen und Dateinamen unverändert lassen.</li><li><code>confluence-storage.xml</code> als technische Vorlage verwenden.</li><li>Inhalt übernehmen und Formatierung prüfen.</li><li><strong>Include-Platzhalter bewusst auf die endgültige Zielseite abbilden.</strong></li><li>Interne Link-Platzhalter nach der finalen Struktur auflösen.</li><li>Medien und Formatierung abschließend prüfen.</li></ol><h2>Seiten</h2><ul>")
    for _, p := range pages { fmt.Fprintf(&b, "<li><code>%s</code> — %s — %d Medien, %d Includes, %d Warnungen</li>", htmlEscape(p.ID), htmlEscape(p.Title), len(p.Media), len(p.Includes), len(p.Warnings)) }
    b.WriteString("</ul><p><strong>Wichtig:</strong> Die zukünftige Confluence-Struktur wird nicht automatisch festgelegt.</p>")
    return b.String()
}

func writeCSV(z *zip.Writer, name string, header []string, rows [][]string) {
    var b strings.Builder; w := csv.NewWriter(&b); _ = w.Write(header); for _, row := range rows { _ = w.Write(row) }; w.Flush(); write(z, name, []byte(b.String()))
}
func write(z *zip.Writer, n string, b []byte) { w, e := z.Create(strings.TrimPrefix(filepath.ToSlash(n), "/")); if e == nil { _, _ = w.Write(b) } }
func htmlEscape(s string) string { return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s) }
