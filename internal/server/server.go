package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/config"
	exporter "github.com/TimUx/Doku-Wiki-Confluence_migration/internal/export"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/render"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/scanner"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/security"
	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/store"
)

//go:embed web/*
var assets embed.FS

type App struct {
	cfg    config.Config
	db     *store.Store
	scanMu sync.Mutex
}

func New(c config.Config, s *store.Store) *App { return &App{cfg: c, db: s} }
func (a *App) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /ready", a.ready)
	m.HandleFunc("GET /api/status", a.status)
	m.HandleFunc("POST /api/scans", a.scan)
	m.HandleFunc("GET /api/pages", a.pages)
	m.HandleFunc("GET /api/pages/{id}", a.page)
	m.HandleFunc("POST /api/exports", a.export)
	sub, _ := fs.Sub(assets, "web")
	m.Handle("/", http.FileServer(http.FS(sub)))
	return securityHeaders(m)
}
func (a *App) ready(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, map[string]string{"status": "ready"})
}
func (a *App) status(w http.ResponseWriter, r *http.Request) { jsonOut(w, a.db.Status()) }
func (a *App) scan(w http.ResponseWriter, r *http.Request) {
	if !a.scanMu.TryLock() {
		http.Error(w, "scan already running", http.StatusConflict)
		return
	}
	defer a.scanMu.Unlock()
	res, e := scanner.Scan(r.Context(), a.cfg.DokuWiki.PagesPath, a.cfg.DokuWiki.MediaPath)
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	if e = a.db.Replace(res.Pages, res.Media, len(res.Warnings)); e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	jsonOut(w, map[string]any{"pages": len(res.Pages), "media": len(res.Media), "warnings": res.Warnings})
}
func (a *App) pages(w http.ResponseWriter, r *http.Request) {
	p, e := a.db.Pages(strings.TrimSpace(r.URL.Query().Get("q")))
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	for i := range p {
		p[i].Source = ""
		p[i].Nodes = nil
	}
	jsonOut(w, p)
}
func (a *App) page(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !security.ValidID(id) {
		http.Error(w, "invalid page id", 400)
		return
	}
	p, e := a.db.Page(id)
	if e != nil {
		http.Error(w, "page not found", 404)
		return
	}
	jsonOut(w, map[string]any{"page": p, "preview": render.HTML(p)})
}
func (a *App) export(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string   `json:"name"`
		Pages []string `json:"pages"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req) != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	for _, id := range req.Pages {
		if !security.ValidID(id) {
			http.Error(w, "invalid page id", 400)
			return
		}
	}
	path, e := exporter.Create(a.db, a.cfg.Storage.ExportDirectory, req.Name, req.Pages)
	if e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	clean, e := security.Within(a.cfg.Storage.ExportDirectory, path)
	if e != nil {
		http.Error(w, "invalid export path", 500)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(clean))
	http.ServeFile(w, r, clean)
}
func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

var _ = os.ErrNotExist
