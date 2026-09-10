package render

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

var bold = regexp.MustCompile(`\*\*(.+?)\*\*`)
var italic = regexp.MustCompile(`//(.+?)//`)

func Inline(s string) string {
	s = html.EscapeString(s)
	s = bold.ReplaceAllString(s, "<strong>$1</strong>")
	s = italic.ReplaceAllString(s, "<em>$1</em>")
	return s
}
func HTML(p model.Page) string {
	var b strings.Builder
	for _, n := range p.Nodes {
		switch n.Type {
		case "heading":
			fmt.Fprintf(&b, "<h%s>%s</h%s>", n.Level, Inline(n.Text), n.Level)
		case "paragraph":
			fmt.Fprintf(&b, "<p>%s</p>", Inline(n.Text))
		case "bullet_item":
			fmt.Fprintf(&b, "<ul><li>%s</li></ul>", Inline(strings.TrimSpace(strings.TrimPrefix(n.Text, "*"))))
		case "ordered_item":
			fmt.Fprintf(&b, "<ol><li>%s</li></ol>", Inline(strings.TrimSpace(strings.TrimPrefix(n.Text, "-"))))
		case "quote":
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>", Inline(strings.TrimPrefix(n.Text, ">")))
		case "code":
			fmt.Fprintf(&b, "<pre><code>%s</code></pre>", html.EscapeString(n.Text))
		case "info", "warning", "note":
			fmt.Fprintf(&b, "<aside class=\"panel %s\"><strong>%s</strong><p>%s</p></aside>", n.Type, strings.ToUpper(n.Type), Inline(n.Text))
		case "include":
			fmt.Fprintf(&b, "<div class=\"include\">Include Page: %s</div>", html.EscapeString(n.Target))
		case "unknown_plugin":
			fmt.Fprintf(&b, "<div class=\"unsupported\"><strong>Unsupported plugin: %s</strong><pre>%s</pre></div>", html.EscapeString(n.Plugin), html.EscapeString(n.Raw))
		}
	}
	return b.String()
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
			fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(strings.TrimPrefix(n.Text, "*")))
		case "ordered_item":
			fmt.Fprintf(&b, "1. %s\n", strings.TrimSpace(strings.TrimPrefix(n.Text, "-")))
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
