package render

import (
	"fmt"
	"html"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var confluenceMedia = regexp.MustCompile(`\{\{\s*([^}|?]+)(?:\?[^}|]*)?(?:\|([^}]*))?\s*\}\}`)
var confluenceLink = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
var confluenceBold = regexp.MustCompile(`\*\*(.+?)\*\*`)
var confluenceItalic = regexp.MustCompile(`//([^/\n]+?)//`)

// AttachmentNames maps DokuWiki media targets to the exact filenames used in the export.
func AttachmentNames(p model.Page) map[string]string {
	out := map[string]string{}
	used := map[string]int{}
	for _, ref := range p.Media {
		if _, ok := out[ref.Target]; ok { continue }
		name := filepath.Base(strings.ReplaceAll(ref.Target, ":", "/"))
		if name == "." || name == "" { name = "attachment" }
		var b strings.Builder
		for _, r := range name {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("._-", r) { b.WriteRune(r) } else { b.WriteByte('_') }
		}
		name = b.String()
		key := strings.ToLower(name)
		used[key]++
		if used[key] > 1 {
			ext := filepath.Ext(name)
			name = strings.TrimSuffix(name, ext) + fmt.Sprintf("_%d", used[key]) + ext
		}
		out[ref.Target] = name
	}
	return out
}

// ConfluenceStorage produces Confluence storage-format XML. Includes are deliberately
// kept as placeholders because their future Confluence locations are unknown.
func ConfluenceStorage(p model.Page) string {
	attachments := AttachmentNames(p)
	var b strings.Builder
	b.WriteString(`<ac:structured-macro ac:name="toc"/>` + "\n")
	for _, n := range p.Nodes {
		switch n.Type {
		case "heading":
			fmt.Fprintf(&b, "<h%s>%s</h%s>\n", n.Level, storageInline(n.Text, p, attachments), n.Level)
		case "paragraph":
			fmt.Fprintf(&b, "<p>%s</p>\n", storageInline(n.Text, p, attachments))
		case "bullet_item":
			fmt.Fprintf(&b, "<ul><li>%s</li></ul>\n", storageInline(strings.TrimSpace(strings.TrimPrefix(n.Text, "*")), p, attachments))
		case "ordered_item":
			fmt.Fprintf(&b, "<ol><li>%s</li></ol>\n", storageInline(strings.TrimSpace(strings.TrimPrefix(n.Text, "-")), p, attachments))
		case "quote":
			fmt.Fprintf(&b, "<blockquote><p>%s</p></blockquote>\n", storageInline(strings.TrimSpace(strings.TrimPrefix(n.Text, ">")), p, attachments))
		case "code":
			code := strings.ReplaceAll(n.Text, "]]>", "]]]]><![CDATA[>")
			fmt.Fprintf(&b, "<ac:structured-macro ac:name=\"code\"><ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body></ac:structured-macro>\n", code)
		case "info", "warning", "note":
			title := strings.ToUpper(n.Type)
			if n.Meta != nil && n.Meta["kind"] == "important" { title = "WICHTIG" }
			fmt.Fprintf(&b, "<ac:structured-macro ac:name=\"panel\"><ac:parameter ac:name=\"title\">%s</ac:parameter><ac:rich-text-body><p>%s</p></ac:rich-text-body></ac:structured-macro>\n", html.EscapeString(title), storageInline(n.Text, p, attachments))
		case "include":
			fmt.Fprintf(&b, "<p>[DOKUWIKI INCLUDE: %s]</p>\n", html.EscapeString(n.Target))
		case "unknown_plugin":
			fmt.Fprintf(&b, "<p><strong>Unsupported plugin: %s</strong></p><pre>%s</pre>\n", html.EscapeString(n.Plugin), html.EscapeString(n.Raw))
		}
	}
	return b.String()
}

func storageInline(s string, p model.Page, attachments map[string]string) string {
	out := html.EscapeString(s)
	out = confluenceMedia.ReplaceAllStringFunc(out, func(raw string) string {
		m := confluenceMedia.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		target := normalizeMediaTarget(resolveTarget(p.ID, m[1]))
		name := attachments[target]
		if name == "" { name = filepath.Base(strings.ReplaceAll(target, ":", "/")) }
		return fmt.Sprintf(`<ac:image><ri:attachment ri:filename="%s"/></ac:image>`, html.EscapeString(name))
	})
	out = confluenceLink.ReplaceAllStringFunc(out, func(raw string) string {
		m := confluenceLink.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		label := m[2]
		if label == "" { label = m[1] }
		if isExternalLink(m[1]) {
			return fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(normalizeURL(m[1])), html.EscapeString(label))
		}
		if strings.HasPrefix(strings.TrimSpace(m[1]), "#") {
			return fmt.Sprintf(`<a href="#%s">%s</a>`, html.EscapeString(anchorID(strings.TrimPrefix(strings.TrimSpace(m[1]), "#"))), html.EscapeString(label))
		}
		resolved := resolveTarget(p.ID, m[1])
		return fmt.Sprintf(`<span>[DOKUWIKI LINK: %s]</span>`, html.EscapeString(resolved))
	})
	out = confluenceBold.ReplaceAllString(out, `<strong>$1</strong>`)
	out = confluenceItalic.ReplaceAllString(out, `<em>$1</em>`)
	return out
}

func resolveTarget(current, target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, ":") { return strings.TrimPrefix(target, ":") }
	if strings.HasPrefix(target, "#") { return target }
	if strings.Contains(target, ":") { return target }
	if i := strings.LastIndex(current, ":"); i >= 0 { return current[:i+1] + target }
	return target
}
