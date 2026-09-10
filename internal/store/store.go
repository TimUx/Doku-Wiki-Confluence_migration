package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS pages(id TEXT PRIMARY KEY, namespace TEXT,title TEXT,source_file TEXT,size INTEGER,modified TEXT,document TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS media(id TEXT PRIMARY KEY,path TEXT,size INTEGER,modified TEXT);
CREATE TABLE IF NOT EXISTS scans(id INTEGER PRIMARY KEY,started TEXT,finished TEXT,pages INTEGER,media INTEGER,warnings INTEGER);
CREATE TABLE IF NOT EXISTS exports(id INTEGER PRIMARY KEY,created TEXT,path TEXT,pages INTEGER);
CREATE TABLE IF NOT EXISTS migration_pages(migration_id INTEGER,page_id TEXT,status TEXT,updated TEXT,PRIMARY KEY(migration_id,page_id));`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Replace(pages []model.Page, media []model.Media, warnings int) error {
	tx, e := s.db.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.Exec("DELETE FROM pages; DELETE FROM media"); e != nil {
		return e
	}
	for _, p := range pages {
		b, _ := json.Marshal(p)
		if _, e = tx.Exec("INSERT INTO pages VALUES(?,?,?,?,?,?,?)", p.ID, p.Namespace, p.Title, p.SourceFile, p.Size, p.Modified.Format(time.RFC3339Nano), b); e != nil {
			return e
		}
	}
	for _, m := range media {
		if _, e = tx.Exec("INSERT INTO media VALUES(?,?,?,?)", m.ID, m.Path, m.Size, m.Modified.Format(time.RFC3339Nano)); e != nil {
			return e
		}
	}
	_, e = tx.Exec("INSERT INTO scans(started,finished,pages,media,warnings) VALUES(?,?,?,?,?)", time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), len(pages), len(media), warnings)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Store) Pages(q string) ([]model.Page, error) {
	query := "SELECT document FROM pages"
	args := []any{}
	if q != "" {
		query += " WHERE id LIKE ? OR title LIKE ? OR namespace LIKE ? OR document LIKE ?"
		v := "%" + q + "%"
		args = []any{v, v, v, v}
	}
	query += " ORDER BY id"
	rows, e := s.db.Query(query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Page
	for rows.Next() {
		var b []byte
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		var p model.Page
		if e = json.Unmarshal(b, &p); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *Store) Page(id string) (model.Page, error) {
	var b []byte
	e := s.db.QueryRow("SELECT document FROM pages WHERE id=?", id).Scan(&b)
	var p model.Page
	if e == nil {
		e = json.Unmarshal(b, &p)
	}
	return p, e
}
func (s *Store) Status() map[string]any {
	var pages, media int
	_ = s.db.QueryRow("SELECT count(*) FROM pages").Scan(&pages)
	_ = s.db.QueryRow("SELECT count(*) FROM media").Scan(&media)
	var scan sql.NullString
	_ = s.db.QueryRow("SELECT finished FROM scans ORDER BY id DESC LIMIT 1").Scan(&scan)
	return map[string]any{"pages": pages, "media": media, "lastScan": scan.String}
}
func (s *Store) MediaPath(id string) (string, error) {
	var p string
	e := s.db.QueryRow("SELECT path FROM media WHERE id=?", id).Scan(&p)
	if e != nil {
		return "", fmt.Errorf("media %s not found", id)
	}
	return p, nil
}
func (s *Store) RecordExport(path string, pages int) {
	_, _ = s.db.Exec("INSERT INTO exports(created,path,pages) VALUES(?,?,?)", time.Now().Format(time.RFC3339), path, pages)
}
