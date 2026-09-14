package render

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var bold = regexp.MustCompile(`\*\*(.+?)\*\*`)
var italic = regexp.MustCompile(`//([^/\n]+?)//`)
var underline = regexp.MustCompile(`__(.+?)__`)
var mono = regexp.MustCompile(`''(.+?)''`)
var strike = regexp.MustCompile(`~~(.+?)~~`)
var sub = regexp.MustCompile(`_(\{[^}]+\}|\w+)`)
var sup = regexp.MustCompile(`\^(\{[^}]+\}|\w+)`)
var footnote = regexp.MustCompile(`\(\((.+?)\)\)`)
var media = regexp.MustCompile(`\{\{\s*([^}|?]+)(?:\?[^}|]*)?(?:\|([^}]*))?\s*\}\}`)
var wikiLink = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)
var wrapInline = regexp.MustCompile(`(?is)<(wrap|inline|span)\b([^>]*)>(.*?)</(wrap|inline|span)>`)
var wrapWidth = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:%|px|em|rem|vw|vh)$`)

func Inline(s string) string {
	s = html.EscapeString(s)
	s = bold.ReplaceAllString(s, "<strong>$1</strong>")
	s = italic.ReplaceAllString(s, "<em>$1</em>")
	s = underline.ReplaceAllString(s, "<u>$1</u>")
	s = mono.ReplaceAllString(s, "<code>$1</code>")
	s = strike.ReplaceAllString(s, "<del>$1</del>")
	s = sub.ReplaceAllString(s, "<sub>$1</sub>")
	s = sup.ReplaceAllString(s, "<sup>$1</sup>")
	return s
}

func previewInline(s, current string) string {
	out := html.EscapeString(s)
	out = wikiLink.ReplaceAllStringFunc(out, func(raw string) string {
		m := wikiLink.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		target := strings.TrimSpace(m[1])
		label := m[2]
		if label == "" { label = target }
		if isExternalLink(target) {
			return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, html.EscapeString(normalizeURL(target)), previewInline(label, current))
		}
		if strings.HasPrefix(target, "#") {
			return fmt.Sprintf(`<a href="#%s">%s</a>`, html.EscapeString(anchorID(strings.TrimPrefix(target, "#"))), previewInline(label, current))
		}
		return fmt.Sprintf(`<span class="internal-link">[DOKUWIKI LINK: %s]</span>`, html.EscapeString(resolveTarget(current, target)))
	})
	out = media.ReplaceAllStringFunc(out, func(raw string) string {
		m := media.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		target := normalizeMediaTarget(resolveTarget(current, m[1]))
		label := m[2]
		if label == "" { label = target }
		return fmt.Sprintf(`<img class="dokuwiki-media" src="/api/media?target=%s" alt="%s" title="%s">`, url.QueryEscape(target), html.EscapeString(label), html.EscapeString(label))
	})
	out = wrapInline.ReplaceAllStringFunc(out, func(raw string) string {
		m := wrapInline.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		return `<span class="dokuwiki-wrap ` + html.EscapeString(strings.TrimSpace(m[2])) + `">` + previewInline(m[3], current) + `</span>`
	})
	out = footnote.ReplaceAllStringFunc(out, func(raw string) string {
		m := footnote.FindStringSubmatch(html.UnescapeString(raw))
		if len(m) == 0 { return raw }
		return `<sup class="footnote">` + previewInline(m[1], current) + `</sup>`
	})
	out = bold.ReplaceAllString(out, "<strong>$1</strong>")
	out = italic.ReplaceAllString(out, "<em>$1</em>")
	out = underline.ReplaceAllString(out, "<u>$1</u>")
	out = mono.ReplaceAllString(out, "<code>$1</code>")
	out = strike.ReplaceAllString(out, "<del>$1</del>")
	out = sub.ReplaceAllString(out, "<sub>$1</sub>")
	out = sup.ReplaceAllString(out, "<sup>$1</sup>")
	return out
}

func nodeIndent(n model.Node) int {
	if n.Meta == nil { return 0 }
	v, err := strconv.Atoi(n.Meta["indent"])
	if err != nil { return 0 }
	return v
}

func normalizeMediaTarget(target string) string {
	target = unescapeDokuWiki(target)
	if i := strings.Index(target, "fetch.php/"); i >= 0 { target = target[i+len("fetch.php/"):] }
	return strings.TrimPrefix(target, "/")
}

func unescapeDokuWiki(s string) string {
	for _, p := range []struct{ a, b string }{{`\:`, ":"}, {`\_`, "_"}, {`\.`, "."}, {`\-`, "-"}, {`\+`, "+"}, {`\#`, "#"}, {`\&`, "&"}, {`\?`, "?"}, {`\|`, "|"}, {`\*`, "*"}} {
		s = strings.ReplaceAll(s, p.a, p.b)
	}
	return s
}

func normalizeURL(s string) string {
	s = unescapeDokuWiki(strings.TrimSpace(s))
	s = strings.Replace(s, "https:*", "https://", 1)
	s = strings.Replace(s, "http:*", "http://", 1)
	return s
}

func isExternalLink(s string) bool {
	s = normalizeURL(s)
	return strings.Contains(s, "://") || strings.HasPrefix(strings.ToLower(s), "mailto:")
}

func anchorID(s string) string {
	s = strings.ToLower(strings.TrimSpace(unescapeDokuWiki(s)))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r >= 128 { b.WriteRune(r); dash = false
		} else if !dash && b.Len() > 0 { b.WriteByte('-'); dash = true }
	}
	return strings.Trim(b.String(), "-")
}

func resolveTarget(current, target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, ":") { return strings.TrimPrefix(target, ":") }
	if strings.HasPrefix(target, "#") { return target }
	if strings.Contains(target, ":") { return target }
	if i := strings.LastIndex(current, ":"); i >= 0 { return current[:i+1] + target }
	return target
}
