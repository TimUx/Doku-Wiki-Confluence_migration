package export

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/parser"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

func TestCreateEndToEnd(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dbPath := filepath.Join(root, "test.db")
	exportDir := filepath.Join(root, "exports")
	mediaDir := filepath.Join(root, "media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil { t.Fatal(err) }

	mediaPath := filepath.Join(mediaDir, "backup-schema.png")
	if err := os.WriteFile(mediaPath, []byte("test-image"), 0644); err != nil { t.Fatal(err) }

	db, err := store.Open(dbPath)
	if err != nil { t.Fatal(err) }
	defer db.Close()

	rootPage := parser.Parse("server", `====== Server ======

Die zentrale [[:server:backup|Backup-Dokumentation]] liegt im Namespace.

{{page>server:backup}}
`)
	rootPage.Namespace = ""
	rootPage.SourceFile = "server.txt"

	backupPage := parser.Parse("server:backup", `====== Backup ======

Das Backup-Schema ist in {{:images:backup-schema.png|Backup Schema}} dokumentiert.

Siehe auch [[server:restore|Restore-Anleitung]].
`)
	backupPage.Namespace = "server"
	backupPage.SourceFile = "server/backup.txt"

	restorePage := parser.Parse("server:restore", `====== Restore ======

Die [[:server:backup|Backup-Seite]] beschreibt die Ausgangslage.

{{page>server:backup}}
`)
	restorePage.Namespace = "server"
	restorePage.SourceFile = "server/restore.txt"

	pages := []model.Page{rootPage, backupPage, restorePage}
	media := []model.Media{{
		ID: "images:backup-schema.png", Path: mediaPath,
		Size: int64(len("test-image")), Modified: time.Now(),
	}}
	if err := db.Replace(pages, media, 0); err != nil { t.Fatal(err) }

	exportPath, err := Create(db, exportDir, "migration", []string{
		"server", "server:backup", "server:restore", "server:backup",
	}, "https://confluence.example.local")
	if err != nil { t.Fatal(err) }

	r, err := zip.OpenReader(exportPath)
	if err != nil { t.Fatal(err) }
	defer r.Close()

	files := make(map[string]string)
	for _, f := range r.File {
		content, err := f.Open()
		if err != nil { t.Fatal(err) }
		data, err := io.ReadAll(content)
		_ = content.Close()
		if err != nil { t.Fatal(err) }
		files[f.Name] = string(data)
	}

	expectedFiles := []string{
		"manifest.json",
		"pages/001_Server/page.html",
		"pages/001_Server/page.md",
		"pages/001_Server/confluence-storage.xml",
		"pages/001_Server/source.txt",
		"pages/001_Server/metadata.json",
		"pages/001_Server/attachments.csv",
		"pages/001_Server/includes.csv",
		"pages/002_Backup/page.html",
		"pages/002_Backup/page.md",
		"pages/002_Backup/confluence-storage.xml",
		"pages/002_Backup/source.txt",
		"pages/002_Backup/metadata.json",
		"pages/002_Backup/attachments/backup-schema.png",
		"pages/002_Backup/attachments.csv",
		"pages/002_Backup/includes.csv",
		"pages/003_Restore/page.html",
		"pages/003_Restore/confluence-storage.xml",
		"pages/003_Restore/includes.csv",
		"MIGRATION-GUIDE.md",
		"MIGRATION-GUIDE.html",
		"migration-report.html",
		"migration-report.json",
	}
	for _, name := range expectedFiles {
		if _, ok := files[name]; !ok { t.Errorf("expected ZIP entry %q", name) }
	}

	var manifest Manifest
	if err := json.Unmarshal([]byte(files["manifest.json"]), &manifest); err != nil { t.Fatal(err) }
	if manifest.Target["baseURL"] != "https://confluence.example.local" { t.Errorf("unexpected Confluence base URL: %q", manifest.Target["baseURL"]) }
	if len(manifest.Pages) != 3 { t.Fatalf("expected 3 pages in manifest, got %d", len(manifest.Pages)) }
	if got := []string{manifest.Pages[0].ID, manifest.Pages[1].ID, manifest.Pages[2].ID}; strings.Join(got, ",") != "server,server:backup,server:restore" {
		t.Errorf("unexpected page order: %v", got)
	}
	if manifest.Pages[1].Namespace != "server" || manifest.Pages[2].Namespace != "server" {
		t.Errorf("namespace hierarchy not preserved in manifest: %+v", manifest.Pages)
	}

	rootStorage := files["pages/001_Server/confluence-storage.xml"]
	if !strings.Contains(rootStorage, `ac:name="include"`) || !strings.Contains(rootStorage, `<ri:page ri:content-title="backup"/>`) {
		t.Errorf("root page include was not rendered as native Confluence Include Page macro: %s", rootStorage)
	}

	backupStorage := files["pages/002_Backup/confluence-storage.xml"]
	if !strings.Contains(backupStorage, `ri:filename="backup-schema.png"`) { t.Errorf("backup page does not reference exported attachment name") }
	if !strings.Contains(backupStorage, "[DOKUWIKI LINK: server:restore]") {
		t.Errorf("backup page lost its internal link placeholder: %s", backupStorage)
	}

	restoreStorage := files["pages/003_Restore/confluence-storage.xml"]
	if !strings.Contains(restoreStorage, "[DOKUWIKI LINK: server:backup]") {
		t.Errorf("restore page lost its cross-page link placeholder: %s", restoreStorage)
	}

	includes := files["pages/001_Server/includes.csv"]
	if !strings.Contains(includes, "server:backup") || !strings.Contains(includes, "[DOKUWIKI INCLUDE: server:backup]") { t.Errorf("root include manifest is incomplete: %s", includes) }

	attachments := files["pages/002_Backup/attachments.csv"]
	if !strings.Contains(attachments, "images:backup-schema.png") || !strings.Contains(attachments, "backup-schema.png") || !strings.Contains(attachments, "image/png") {
		t.Errorf("attachment manifest does not contain the expected mapping: %s", attachments)
	}

	if got := files["pages/002_Backup/attachments/backup-schema.png"]; got != "test-image" { t.Errorf("exported attachment content mismatch: %q", got) }

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "pages/004_") { t.Errorf("duplicate include target was unexpectedly exported: %s", f.Name) }
	}
}
