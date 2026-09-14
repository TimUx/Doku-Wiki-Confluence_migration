package render

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var bold = regexp.MustCompile(`\*\*(.+?)\*\*`)
var italic = regexp.MustCompile(`//(.+?)//`)
var media = regexp.MustCompile(`\{\{\s*([^}|?]+)(?:\?[^}|]*)?(?:\|([^}]+))?\s*\}\}`)

func Inline(s string) string {
	s = html.EscapeString(s)
	s = bold.ReplaceAllString(s, "<strong>$1</strong>")
	s = italic.ReplaceAllString(s, "<em>$1</em>")
	return s
}

func HTML(p model.Page) string {
	var b strings.Builder
	for i := 0; i < len(p.Nodes); i++ {
		n := p.Nodes[i]
		switch n.Type {
		case "heading":
			fmt.Fprintf(&b, "<h%s>%s</h%s>", n.Level, Inline(n.Text), n.Level)
		case "paragraph":
			fmt.Fprintf(&b, "<p>%s</p>", Inline(n.Text))
		case "bullet_item", "ordered_item":
			ordered := n.Type == "ordered_item"
			tag := "ul"
			if ordered {
				tag = "ol"
			}
			fmt.Fprintf(&b, "<%s>", tag)
			for i < len(p.Nodes) && p.Nodes[i].Type == n.Type {
				fmt.Fprintf(&b, "<li>%s</li>", Inline(p.Nodes[i].Text))
				i++
			}
			b.WriteString("</" + tag + ">")
			i--
		case "table_row":
			b.WriteString(`<table class="dokuwiki-table"><tbody>`)
			for i < len(p.Nodes) && p.Nodes[i].Type == "table_row" {
				b.WriteString("<tr>")
				for _, cell := range p.Nodes[i].Children {
					tag := "td"
					if cell.Type == "table_header" {
						tag = "th"
					}
					fmt.Fprintf(&b, "<%s>%s</%s>", tag, Inline(cell.Text), tag)
				}
				b.WriteString("</tr>")
				i++
			}
			b.WriteString("</tbody></table>")
			i--
		case "quote":
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>", Inline(n.Text))
		case "code":
			fmt.Fprintf(&b, "<pre><code>%s</code></pre>", html.EscapeString(n.Text))
		case "info", "warning", "note":
			class := "hint " + n.Type
			title := strings.ToUpper(n.Type)
			if n.Meta != nil && n.Meta["kind"] == "important" {
				title = "WICHTIG"
			}
			b.WriteString(`<aside class="` + class + `"><div class="hint-title">` + title + `</div><div class="hint-body">`)
			for _, line := range strings.Split(strings.TrimSpace(n.Text), "\n") {
				if strings.TrimSpace(line) != "" {
					fmt.Fprintf(&b, "<p>%s</p>", Inline(strings.TrimSpace(line)))
				}
			}
			b.WriteString("</div></aside>")
		case "include":
			fmt.Fprintf(&b, "<div class=\"include\">Include Page: %s</div>", html.EscapeString(n.Target))
		case "unknown_plugin":
			fmt.Fprintf(&b, "<div class=\"unsupported\"><strong>Unsupported plugin: %s</strong><pre>%s</pre></div>", html.EscapeString(n.Plugin), html.EscapeString(n.Raw))
		}
	}
	return b.String()
}

// PreviewHTML is the browser preview variant of HTML. It resolves DokuWiki
// media references to the application's read-only media endpoint while
// keeping the export HTML independent from the running application.
func PreviewHTML(p model.Page) string {
	out := HTML(p)
	return media.ReplaceAllStringFunc(out, func(raw string) string {
		m := media.FindStringSubmatch(raw)
		if len(m) == 0 {
			return raw
		}
		target := resolveTarget(p.ID, m[1])
		label := m[2]
		if label == "" {
			label = target
		}
		return fmt.Sprintf(`<img class="dokuwiki-media" src="/api/media?target=%s" alt="%s" title="%s">`, url.QueryEscape(target), html.EscapeString(label), html.EscapeString(label))
	})
}

func resolveTarget(current, target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, ":") {
		return strings.TrimPrefix(target, ":")
	}
	if strings.Contains(target, ":") {
		return target
	}
	if i := strings.LastIndex(current, ":"); i >= 0 {
		return current[:i+1] + target
	}
	return target
}

func Markdown(p model.Page) string {
	var b strings.Builder
	for _, n := range p.Nodes {
		switch n.Type {
		case "heading":
			lvl := n.Level
			if lvl == "" {
				lvl = "2"
			}
			fmt.Fprintf(&b, "%s %s\n\n", strings.Repeat("#", int(lvl[0]-'0')), n.Text)
		case "bullet_item":
			fmt.Fprintf(&b, "- %s\n", n.Text)
		case "ordered_item":
			fmt.Fprintf(&b, "1. %s\n", n.Text)
		case "table_row":
			b.WriteString("| ")
			for i, cell := range n.Children {
				if i > 0 {
					b.WriteString(" | ")
				}
				b.WriteString(cell.Text)
			}
			b.WriteString(" |\n")
		case "code":
			fmt.Fprintf(&b, "```\n%s\n```\n\n", n.Text)
		case "include":
			fmt.Fprintf(&b, "> Include Page: `%s`\n\n", n.Target)
		case "unknown_plugin":
			fmt.Fprintf(&b, "> **Unsupported plugin `%s`**\n> `%s`\n\n", n.Plugin, n.Raw)
		default:
			fmt.Fprintf(&b, "%s\n\n", n.Text)
		}
	}
	return b.String()
}
