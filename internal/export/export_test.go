package export

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

func TestCreateEndToEnd(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	exportDir := filepath.Join(root, "exports")
	mediaDir := filepath.Join(root, "media")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		t.Fatal(err)
	}

	/*
	 * Store mit einer Testseite und einem Testmedium
	 * initialisieren.
	 *
	 * Die konkrete Initialisierung muss an die bestehende
	 * store.Store API angepasst werden.
	 */

	page := model.Page{
		ID:         "server:backup",
		Namespace:  "server",
		Title:      "Backup",
		SourceFile: "server/backup.txt",
		Source: `
====== Backup ======

Ein Test für den Export.

{{:images:backup-schema.png|Backup Schema}}

{{page>server:restore}}
`,
	}

	page.Media = []model.Reference{
		{
			Target:  "images:backup-schema.png",
			Display: "Backup Schema",
			Kind:    "image",
		},
	}

	page.Includes = []model.Reference{
		{
			Target: "server:restore",
			Kind:   "include",
		},
	}

	/*
	 * Test-Mediendatei.
	 */
	mediaPath := filepath.Join(
		mediaDir,
		"backup-schema.png",
	)

	if err := os.WriteFile(
		mediaPath,
		[]byte("test-image"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	/*
	 * Store befüllen.
	 *
	 * Hier muss die konkrete bestehende Store-API verwendet werden.
	 */

	_ = page
	_ = mediaPath

	/*
	 * Export durchführen.
	 */

	// exportPath, err := Create(db, exportDir, "migration", []string{page.ID})
	// if err != nil {
	//     t.Fatal(err)
	// }

	/*
	 * ZIP öffnen und Inhalte prüfen.
	 */

	// r, err := zip.OpenReader(exportPath)
	// if err != nil {
	//     t.Fatal(err)
	// }
	// defer r.Close()

	expectedFiles := []string{
		"manifest.json",
		"MIGRATION-GUIDE.md",
		"MIGRATION-GUIDE.html",
		"migration-report.html",
		"migration-report.json",
	}

	_ = expectedFiles
}
