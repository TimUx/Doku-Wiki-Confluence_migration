package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct { Listen string `yaml:"listen"`; BaseURL string `yaml:"base_url"` } `yaml:"server"`
	DokuWiki struct { Root string `yaml:"root"`; PagesPath string `yaml:"pages_path"`; MediaPath string `yaml:"media_path"`; BaseURL string `yaml:"base_url"` } `yaml:"dokuwiki"`
	Storage struct { Database string `yaml:"database"`; WorkDirectory string `yaml:"work_directory"`; ExportDirectory string `yaml:"export_directory"` } `yaml:"storage"`
	Export struct { IncludeMode string `yaml:"include_mode"`; PreserveOriginalSource bool `yaml:"preserve_original_source"`; IncludeMarkdown bool `yaml:"include_markdown"`; IncludeHTML bool `yaml:"include_html"`; IncludeMedia bool `yaml:"include_media"` } `yaml:"export"`
	Security struct { ReadOnlySource bool `yaml:"read_only_source"` } `yaml:"security"`
	Logging struct { Level string `yaml:"level"` } `yaml:"logging"`
}

func Load(path string) (Config, error) {
	var c Config; b, err := os.ReadFile(path); if err != nil { return c, fmt.Errorf("read config: %w", err) }
	dec := yaml.NewDecoder(bytesReader(b)); dec.KnownFields(true)
	if err = dec.Decode(&c); err != nil { return c, fmt.Errorf("parse config: %w", err) }
	if c.Server.Listen == "" { c.Server.Listen = "0.0.0.0:8080" }
	if c.Export.IncludeMode == "" { c.Export.IncludeMode = "dynamic" }
	if !c.Security.ReadOnlySource { return c, fmt.Errorf("security.read_only_source must be true") }
	for name, p := range map[string]string{"dokuwiki.pages_path":c.DokuWiki.PagesPath,"dokuwiki.media_path":c.DokuWiki.MediaPath} {
		st, e := os.Stat(p); if e != nil || !st.IsDir() { return c, fmt.Errorf("%s does not exist or is not a directory: %s", name, p) }
	}
	if c.Storage.WorkDirectory == "" || c.Storage.Database == "" || c.Storage.ExportDirectory == "" { return c, fmt.Errorf("storage paths are required") }
	for _, p := range []string{c.Storage.WorkDirectory, filepath.Dir(c.Storage.Database), c.Storage.ExportDirectory} { if err := os.MkdirAll(p, 0750); err != nil { return c, err } }
	if c.Export.IncludeMode != "dynamic" && c.Export.IncludeMode != "static" { return c, fmt.Errorf("export.include_mode must be dynamic or static") }
	return c, nil
}

type reader struct{ b []byte; i int }
func bytesReader(b []byte) *reader { return &reader{b:b} }
func (r *reader) Read(p []byte)(int,error){ if r.i>=len(r.b){return 0,io.EOF}; n:=copy(p,r.b[r.i:]);r.i+=n;return n,nil }
