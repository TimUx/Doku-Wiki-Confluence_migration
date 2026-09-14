package export

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

func TestCreateEndToEnd(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(root, "test.db")
	exportDir := filepath.Join(root, "exports")
	mediaPath := filepath.Join(root, "backup-schema.png")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(mediaPath, []byte("test-image"), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	page := model.Page{
		ID:         "server:backup",
		Namespace:  "server",
		Title:      "Backup",
		SourceFile: "server/backup.txt",
		Source:     "====== Backup ======\n\nEin Test für den Export.\n\n{{:images:backup-schema.png|Backup Schema}}\n\n{{page>server:restore}}\n",
		Modified:   time.Now(),
		Nodes: []model.Node{
			{Type: "heading", Text: "Backup", Level: "1"},
			{Type: "paragraph", Text: "Ein Test für den Export."},
			{Type: "paragraph", Text: "{{:images:backup-schema.png|Backup Schema}}"},
			{Type: "include", Target: "server:restore"},
		},
		Media: []model.Reference{
			{Target: "images:backup-schema.png", Display: "Backup Schema", Kind: "image"},
		},
		Includes: []model.Reference{
			{Target: "server:restore", Kind: "include"},
		},
	}

	media := model.Media{
		ID:       "images:backup-schema.png",
		Path:     mediaPath,
		Size:     int64(len("test-image")),
		Modified: time.Now(),
	}

	if err := db.Replace([]model.Page{page}, []model.Media{media}, 0); err != nil {
		t.Fatal(err)
	}

	exportPath, err := Create(db, exportDir, "migration", []string{page.ID}, "https://confluence.example.local")
	if err != nil {
		t.Fatal(err)
	}

	r, err := zip.OpenReader(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	files := make(map[string]string)
	for _, f := range r.File {
		content, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		data, err := io.ReadAll(content)
		_ = content.Close()
		if err != nil {
			t.Fatal(err)
		}

		files[f.Name] = string(data)
	}

	expectedFiles := []string{
		"manifest.json",
		"pages/001_Backup/page.html",
		"pages/001_Backup/page.md",
		"pages/001_Backup/confluence-storage.xml",
		"pages/001_Backup/source.txt",
		"pages/001_Backup/metadata.json",
		"pages/001_Backup/attachments/backup-schema.png",
		"pages/001_Backup/attachments.csv",
		"pages/001_Backup/includes.csv",
		"MIGRATION-GUIDE.md",
		"MIGRATION-GUIDE.html",
		"migration-report.html",
		"migration-report.json",
	}

	for _, name := range expectedFiles {
		if _, ok := files[name]; !ok {
			t.Errorf("expected ZIP entry %q", name)
		}
	}

	storage := files["pages/001_Backup/confluence-storage.xml"]
	if !strings.Contains(storage, `ri:filename="backup-schema.png"`) {
		t.Errorf("Confluence storage format does not reference the exported attachment name")
	}

	if !strings.Contains(storage, "[DOKUWIKI INCLUDE: server:restore]") {
		t.Errorf("Confluence storage format does not preserve the include placeholder")
	}

	includes := files["pages/001_Backup/includes.csv"]
	if !strings.Contains(includes, "server:restore") || !strings.Contains(includes, "[DOKUWIKI INCLUDE: server:restore]") {
		t.Errorf("include manifest does not contain the expected placeholder")
	}

	attachments := files["pages/001_Backup/attachments.csv"]
	if !strings.Contains(attachments, "images:backup-schema.png") || !strings.Contains(attachments, "backup-schema.png") || !strings.Contains(attachments, "image/png") {
		t.Errorf("attachment manifest does not contain the expected mapping")
	}

	if strings.Contains(storage, "ac:name=\"include\"") {
		t.Errorf("Confluence storage output unexpectedly created an Include Page macro")
	}

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "pages/002_") {
			t.Errorf("include target was unexpectedly exported as an additional page: %s", f.Name)
		}
	}
}
