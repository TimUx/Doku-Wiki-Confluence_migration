package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var (
	heading    = regexp.MustCompile(`^(={2,6})\s*(.*?)\s*(={2,6})\s*$`)
	link       = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
	media      = regexp.MustCompile(`\{\{\s*([^}|?]+)(?:\?[^}|]*)?(?:\|([^}]+))?\s*\}\}`)
	include    = regexp.MustCompile(`(?i)\{\{(?:page|section|include)>\s*([^}&]+)(?:&([^}]+))?\}\}`)
	pluginOpen = regexp.MustCompile(`(?i)^<([a-z][\w-]*)(?:\s+([^>]*))?>\s*$`)
)

func Parse(id, src string) model.Page {
	p := model.Page{ID: id, Source: src, Title: last(id)}
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trim := strings.TrimSpace(raw)
		if trim == "" {
			continue
		}
		if m := heading.FindStringSubmatch(trim); m != nil && len(m[1]) == len(m[3]) {
			lvl := 7 - len(m[1])
			p.Nodes = append(p.Nodes, model.Node{Type: "heading", Text: m[2], Level: strconv.Itoa(lvl)})
			if p.Title == last(id) {
				p.Title = m[2]
			}
			continue
		}
		if strings.HasPrefix(trim, "<code") || strings.HasPrefix(trim, "<file") {
			end := "</code>"
			if strings.HasPrefix(trim, "<file") {
				end = "</file>"
			}
			var b []string
			i++
			for i < len(lines) && !strings.Contains(lines[i], end) {
				b = append(b, lines[i])
				i++
			}
			p.Nodes = append(p.Nodes, model.Node{Type: "code", Text: strings.Join(b, "\n")})
			continue
		}
		if strings.HasPrefix(strings.ToUpper(trim), "<WRAP ") {
			kind := strings.TrimSuffix(strings.TrimPrefix(trim, "<WRAP "), ">")
			var b []string
			i++
			for i < len(lines) && !strings.EqualFold(strings.TrimSpace(lines[i]), "</WRAP>") {
				b = append(b, lines[i])
				i++
			}
			typ := "note"
			k := strings.ToLower(kind)
			if strings.Contains(k, "warning") {
				typ = "warning"
			} else if strings.Contains(k, "info") || strings.Contains(k, "tip") {
				typ = "info"
			}
			p.Plugins = appendUnique(p.Plugins, "wrap")
			p.Nodes = append(p.Nodes, model.Node{Type: typ, Text: strings.Join(b, "\n"), Plugin: "wrap"})
			continue
		}
		if m := include.FindStringSubmatch(trim); m != nil {
			target := resolve(id, m[1])
			p.Includes = append(p.Includes, model.Reference{Target: target, Kind: "include"})
			p.Plugins = appendUnique(p.Plugins, "include")
			p.Nodes = append(p.Nodes, model.Node{Type: "include", Target: target, Raw: trim})
			continue
		}
		if m := pluginOpen.FindStringSubmatch(trim); m != nil {
			name := strings.ToLower(m[1])
			if name != "code" && name != "file" {
				p.Plugins = appendUnique(p.Plugins, name)
				p.Warnings = append(p.Warnings, model.Warning{Code: "unknown_plugin", Message: "Unsupported plugin: " + name, Raw: trim, Line: i + 1})
				p.Nodes = append(p.Nodes, model.Node{Type: "unknown_plugin", Plugin: name, Raw: trim})
				continue
			}
		}
		for _, m := range link.FindAllStringSubmatch(raw, -1) {
			kind := "internal"
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				kind = "external"
			} else {
				target = resolve(id, target)
			}
			p.Links = append(p.Links, model.Reference{Target: target, Display: m[2], Kind: kind})
		}
		for _, m := range media.FindAllStringSubmatch(raw, -1) {
			p.Media = append(p.Media, model.Reference{Target: resolve(id, m[1]), Display: m[2], Kind: "media"})
		}
		typ := "paragraph"
		if strings.HasPrefix(trim, "  *") {
			typ = "bullet_item"
		} else if strings.HasPrefix(trim, "  -") {
			typ = "ordered_item"
		} else if strings.HasPrefix(trim, ">") {
			typ = "quote"
		}
		p.Nodes = append(p.Nodes, model.Node{Type: typ, Text: trim})
	}
	return p
}
func resolve(current, target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, ":") {
		return strings.TrimPrefix(target, ":")
	}
	if strings.Contains(target, ":") {
		return target
	}
	i := strings.LastIndex(current, ":")
	if i < 0 {
		return target
	}
	return current[:i+1] + target
}
func last(id string) string {
	if i := strings.LastIndex(id, ":"); i >= 0 {
		return id[i+1:]
	}
	return id
}
func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}
