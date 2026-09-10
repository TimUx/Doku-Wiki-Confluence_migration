package export

import (
	"archive/zip"
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
	if len(ids) == 0 {
		return "", fmt.Errorf("select at least one page")
	}
	name = safe.ReplaceAllString(name, "-")
	if name == "" {
		name = "migration"
	}
	path := filepath.Join(dir, name+".zip")
	f, e := os.Create(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	var pages []model.Page
	seen := map[string]bool{}
	var add func(string) error
	add = func(id string) error {
		if seen[id] {
			return nil
		}
		seen[id] = true
		p, e := db.Page(id)
		if e != nil {
			return e
		}
		pages = append(pages, p)
		for _, x := range p.Includes {
			if e := add(x.Target); e != nil {
				p.Warnings = append(p.Warnings, model.Warning{Code: "missing_include", Message: e.Error()})
			}
		}
		return nil
	}
	for _, id := range ids {
		if e := add(id); e != nil {
			return "", e
		}
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].ID < pages[j].ID })
	m := Manifest{1, time.Now().Format(time.RFC3339), map[string]string{"type": "dokuwiki"}, map[string]string{"type": "confluence", "version": "10.2.17"}, pages}
	b, _ := json.MarshalIndent(m, "", "  ")
	write(zw, "manifest.json", b)
	for i, p := range pages {
		base := fmt.Sprintf("pages/%03d_%s/", i+1, safe.ReplaceAllString(p.Title, "_"))
		write(zw, base+"page.html", []byte(render.HTML(p)))
		write(zw, base+"page.md", []byte(render.Markdown(p)))
		write(zw, base+"source.txt", []byte(p.Source))
		meta, _ := json.MarshalIndent(p, "", "  ")
		write(zw, base+"metadata.json", meta)
		for _, m := range p.Media {
			src, e := db.MediaPath(m.Target)
			if e != nil {
				continue
			}
			in, e := os.Open(src)
			if e != nil {
				continue
			}
			w, e := zw.Create(base + "attachments/" + safe.ReplaceAllString(filepath.Base(src), "_"))
			if e == nil {
				_, _ = io.Copy(w, in)
			}
			in.Close()
		}
	}
	report := fmt.Sprintf("<!doctype html><meta charset=utf-8><title>Migration report</title><h1>Migration Summary</h1><p>Pages: %d</p><p>Generated: %s</p>", len(pages), htmlEscape(time.Now().Format(time.RFC3339)))
	write(zw, "migration-report.html", []byte(report))
	rb, _ := json.MarshalIndent(map[string]any{"pages": len(pages), "generated": time.Now()}, "", "  ")
	write(zw, "migration-report.json", rb)
	if e = zw.Close(); e != nil {
		return "", e
	}
	db.RecordExport(path, len(pages))
	return path, nil
}
func write(z *zip.Writer, n string, b []byte) {
	w, e := z.Create(strings.TrimPrefix(filepath.ToSlash(n), "/"))
	if e == nil {
		_, _ = w.Write(b)
	}
}
func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}
