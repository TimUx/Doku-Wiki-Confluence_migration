package export

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime"
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

// Create creates a ZIP migration package for the selected DokuWiki pages.
//
// The export deliberately does not resolve DokuWiki includes to future
// Confluence locations. Includes are exported as explicit placeholders and
// documented in includes.csv.
//
// Media files are exported using deterministic attachment names. The same
// names are used by the generated Confluence storage-format XML and recorded
// in attachments.csv.
func Create(db *store.Store, dir, name string, ids []string) (string, error) {
	if len(ids) == 0 {
		return "", fmt.Errorf("select at least one page")
	}

	name = safe.ReplaceAllString(name, "-")
	if name == "" {
		name = "migration"
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}

	path := filepath.Join(dir, name+".zip")

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create export archive: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	var pages []model.Page
	seen := make(map[string]bool)

	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true

		p, err := db.Page(id)
		if err != nil {
			_ = zw.Close()
			_ = f.Close()
			_ = os.Remove(path)

			return "", fmt.Errorf("load page %q: %w", id, err)
		}

		pages = append(pages, p)
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].ID < pages[j].ID
	})

	manifest := Manifest{
		ExportVersion: 1,
		Created:       time.Now().Format(time.RFC3339),
		Source: map[string]string{
			"type": "dokuwiki",
		},
		Target: map[string]string{
			"type":    "confluence",
			"version": "10.2.17",
		},
		Pages: pages,
	}

	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = zw.Close()
		_ = f.Close()
		_ = os.Remove(path)

		return "", fmt.Errorf("create manifest: %w", err)
	}

	if err := write(zw, "manifest.json", b); err != nil {
		_ = zw.Close()
		_ = f.Close()
		_ = os.Remove(path)

		return "", fmt.Errorf("write manifest: %w", err)
	}

	for i, p := range pages {
		base := fmt.Sprintf(
			"pages/%03d_%s/",
			i+1,
			safe.ReplaceAllString(p.Title, "_"),
		)

		names := render.AttachmentNames(p)

		if err := write(
			zw,
			base+"page.html",
			[]byte(render.HTML(p)),
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write HTML for page %q: %w",
				p.ID,
				err,
			))
		}

		if err := write(
			zw,
			base+"page.md",
			[]byte(render.Markdown(p)),
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write Markdown for page %q: %w",
				p.ID,
				err,
			))
		}

		if err := write(
			zw,
			base+"confluence-storage.xml",
			[]byte(render.ConfluenceStorage(p)),
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write Confluence storage format for page %q: %w",
				p.ID,
				err,
			))
		}

		if err := write(
			zw,
			base+"source.txt",
			[]byte(p.Source),
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write source for page %q: %w",
				p.ID,
				err,
			))
		}

		meta, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"create metadata for page %q: %w",
				p.ID,
				err,
			))
		}

		if err := write(
			zw,
			base+"metadata.json",
			meta,
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write metadata for page %q: %w",
				p.ID,
				err,
			))
		}

		/*
		 * Export media attachments.
		 *
		 * The filename used here must be exactly the filename referenced
		 * from confluence-storage.xml. render.AttachmentNames() provides
		 * the deterministic mapping and also handles collisions.
		 */
		var attachments [][]string

		for _, m := range p.Media {
			src, err := db.MediaPath(m.Target)
			if err != nil {
				return cleanupExport(path, zw, f, fmt.Errorf(
					"resolve media %q for page %q: %w",
					m.Target,
					p.ID,
					err,
				))
			}

			in, err := os.Open(src)
			if err != nil {
				return cleanupExport(path, zw, f, fmt.Errorf(
					"open media %q for page %q: %w",
					m.Target,
					p.ID,
					err,
				))
			}

			filename := names[m.Target]
			if filename == "" {
				filename = filepath.Base(src)
			}

			attachmentPath := base + "attachments/" + filename

			w, err := zw.Create(attachmentPath)
			if err != nil {
				in.Close()

				return cleanupExport(path, zw, f, fmt.Errorf(
					"create attachment %q for page %q: %w",
					filename,
					p.ID,
					err,
				))
			}

			_, copyErr := io.Copy(w, in)
			closeErr := in.Close()

			if copyErr != nil {
				return cleanupExport(path, zw, f, fmt.Errorf(
					"copy attachment %q for page %q: %w",
					filename,
					p.ID,
					copyErr,
				))
			}

			if closeErr != nil {
				return cleanupExport(path, zw, f, fmt.Errorf(
					"close media %q for page %q: %w",
					m.Target,
					p.ID,
					closeErr,
				))
			}

			/*
			 * model.Reference intentionally does not contain a MIME field.
			 * Determine the MIME type from the actual exported filename.
			 */
			mimeType := mime.TypeByExtension(
				strings.ToLower(filepath.Ext(filename)),
			)

			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			attachments = append(
				attachments,
				[]string{
					m.Target,
					filename,
					mimeType,
				},
			)
		}

		if err := writeCSV(
			zw,
			base+"attachments.csv",
			[]string{
				"DokuWiki-Ziel",
				"Confluence-Anhangsname",
				"MIME",
			},
			attachments,
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write attachment manifest for page %q: %w",
				p.ID,
				err,
			))
		}

		/*
		 * Includes deliberately remain unresolved.
		 *
		 * The future Confluence page hierarchy is not known during export,
		 * therefore we preserve the original target and position by using
		 * an explicit placeholder.
		 */
		var includes [][]string

		for _, inc := range p.Includes {
			includes = append(
				includes,
				[]string{
					p.ID,
					inc.Target,
					"[DOKUWIKI INCLUDE: " + inc.Target + "]",
				},
			)
		}

		if err := writeCSV(
			zw,
			base+"includes.csv",
			[]string{
				"Quellseite",
				"DokuWiki-Include",
				"Platzhalter",
			},
			includes,
		); err != nil {
			return cleanupExport(path, zw, f, fmt.Errorf(
				"write include manifest for page %q: %w",
				p.ID,
				err,
			))
		}
	}

	if err := write(
		zw,
		"MIGRATION-GUIDE.md",
		[]byte(guideMarkdown(pages)),
	); err != nil {
		return cleanupExport(path, zw, f, fmt.Errorf(
			"write migration guide: %w",
			err,
		))
	}

	if err := write(
		zw,
		"MIGRATION-GUIDE.html",
		[]byte(guideHTML(pages)),
	); err != nil {
		return cleanupExport(path, zw, f, fmt.Errorf(
			"write HTML migration guide: %w",
			err,
		))
	}

	generated := time.Now().Format(time.RFC3339)

	report := fmt.Sprintf(
		"<!doctype html>"+
			"<meta charset=utf-8>"+
			"<title>Migration report</title>"+
			"<h1>Migration Summary</h1>"+
			"<p>Pages: %d</p>"+
			"<p>Includes bleiben als Platzhalter erhalten.</p>"+
			"<p>Bilder werden mit den exportierten Anhangsnamen referenziert.</p>"+
			"<p>Generated: %s</p>",
		len(pages),
		htmlEscape(generated),
	)

	if err := write(
		zw,
		"migration-report.html",
		[]byte(report),
	); err != nil {
		return cleanupExport(path, zw, f, fmt.Errorf(
			"write HTML migration report: %w",
			err,
		))
	}

	rb, err := json.MarshalIndent(
		map[string]any{
			"pages":     len(pages),
			"generated": generated,
			"includes":  "placeholder",
			"attachments": "confluence-reference",
		},
		"",
		"  ",
	)
	if err != nil {
		return cleanupExport(path, zw, f, fmt.Errorf(
			"create migration report: %w",
			err,
		))
	}

	if err := write(
		zw,
		"migration-report.json",
		rb,
	); err != nil {
		return cleanupExport(path, zw, f, fmt.Errorf(
			"write JSON migration report: %w",
			err,
		))
	}

	if err := zw.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(path)

		return "", fmt.Errorf("finalize export archive: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(path)

		return "", fmt.Errorf("close export archive: %w", err)
	}

	if err := db.RecordExport(path, len(pages)); err != nil {
		return "", fmt.Errorf(
			"record export %q: %w",
			path,
			err,
		)
	}

	return path, nil
}

func guideMarkdown(pages []model.Page) string {
	var b strings.Builder

	b.WriteString("# DokuWiki → Confluence 10.2 Migration Guide\n\n")

	b.WriteString(
		"## Grundprinzip\n\n" +
			"Die Exportdateien sind für eine kontrollierte manuelle Übernahme " +
			"vorbereitet. Die zukünftige Confluence-Struktur wird nicht geraten. " +
			"Includes bleiben deshalb als `[DOKUWIKI INCLUDE: ...]` markiert. " +
			"Bilder verweisen bereits auf den exakten exportierten Anhangsnamen.\n\n",
	)

	b.WriteString(
		"## Ablauf\n\n" +
			"1. Zielseite in Confluence anlegen und Titel prüfen.\n" +
			"2. Alle Dateien aus dem jeweiligen `attachments/`-Ordner hochladen. " +
			"Dateinamen **nicht ändern**.\n" +
			"3. `confluence-storage.xml` als technische Referenz für die " +
			"vorbereitete Confluence-Struktur verwenden.\n" +
			"4. Inhalt in die Zielseite übernehmen und Formatierung prüfen.\n" +
			"5. Jeden Include-Platzhalter bewusst durch die später passende " +
			"Confluence-Referenz bzw. das Include Page Macro ersetzen.\n" +
			"6. Interne Link-Platzhalter anhand der finalen Seitenstruktur auflösen.\n" +
			"7. Bilder, Tabellen, Codeblöcke und Links in der Zielseite kontrollieren.\n\n",
	)

	b.WriteString("## Exportierte Seiten\n\n")

	for _, p := range pages {
		fmt.Fprintf(
			&b,
			"- `%s` — %s — %d Medien, %d Includes, %d Warnungen\n",
			p.ID,
			p.Title,
			len(p.Media),
			len(p.Includes),
			len(p.Warnings),
		)
	}

	b.WriteString(
		"\n## Wichtig\n\n" +
			"Der Export entscheidet nicht, wo eine Seite künftig in Confluence " +
			"liegt. Diese Zuordnung erfolgt bewusst erst bei der Übernahme.\n",
	)

	return b.String()
}

func guideHTML(pages []model.Page) string {
	var b strings.Builder

	b.WriteString(
		"<!doctype html>" +
			"<meta charset=utf-8>" +
			"<title>DokuWiki → Confluence Migration Guide</title>" +
			"<h1>DokuWiki → Confluence 10.2</h1>" +
			"<h2>Übernahme</h2>" +
			"<ol>" +
			"<li>Zielseite anlegen.</li>" +
			"<li>Alle Dateien aus <code>attachments/</code> hochladen und " +
			"Dateinamen unverändert lassen.</li>" +
			"<li><code>confluence-storage.xml</code> als technische Vorlage verwenden.</li>" +
			"<li>Inhalt übernehmen und Formatierung prüfen.</li>" +
			"<li><strong>Include-Platzhalter bewusst auf die endgültige " +
			"Zielseite abbilden.</strong></li>" +
			"<li>Interne Link-Platzhalter nach der finalen Struktur auflösen.</li>" +
			"<li>Medien und Formatierung abschließend prüfen.</li>" +
			"</ol>" +
			"<h2>Seiten</h2>" +
			"<ul>",
	)

	for _, p := range pages {
		fmt.Fprintf(
			&b,
			"<li><code>%s</code> — %s — %d Medien, %d Includes, %d Warnungen</li>",
			htmlEscape(p.ID),
			htmlEscape(p.Title),
			len(p.Media),
			len(p.Includes),
			len(p.Warnings),
		)
	}

	b.WriteString(
		"</ul>" +
			"<p><strong>Wichtig:</strong> Die zukünftige Confluence-Struktur " +
			"wird nicht automatisch festgelegt.</p>",
	)

	return b.String()
}

func writeCSV(
	z *zip.Writer,
	name string,
	header []string,
	rows [][]string,
) error {
	var b strings.Builder

	w := csv.NewWriter(&b)

	if err := w.Write(header); err != nil {
		return err
	}

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}

	return write(z, name, []byte(b.String()))
}

func write(z *zip.Writer, name string, data []byte) error {
	cleanName := strings.TrimPrefix(
		filepath.ToSlash(name),
		"/",
	)

	w, err := z.Create(cleanName)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func htmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(s)
}

func cleanupExport(
	path string,
	zw *zip.Writer,
	f *os.File,
	err error,
) (string, error) {
	_ = zw.Close()
	_ = f.Close()
	_ = os.Remove(path)

	return "", err
}
