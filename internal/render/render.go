package render

import (
	"fmt"
	"html"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

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
			fmt.Fprint(&b, renderListHTML(p.Nodes, &i, false, p.ID))
		case "table_row":
			renderTableHTML(&b, p.Nodes, &i, false, p.ID)
		case "quote":
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>", Inline(n.Text))
		case "hr":
			b.WriteString("<hr>")
		case "code":
			fmt.Fprintf(&b, "<pre><code>%s</code></pre>", html.EscapeString(n.Text))
		case "info", "warning", "note":
			fmt.Fprintf(&b, `<aside class="hint %s"><strong>%s</strong><p>%s</p></aside>`, n.Type, html.EscapeString(strings.ToUpper(n.Type)), Inline(n.Text))
		case "wrap":
			renderWrapHTML(&b, n)
		case "include":
			fmt.Fprintf(&b, `<div class="include">Include %s: %s</div>`, html.EscapeString(n.Meta["mode"]), html.EscapeString(n.Target))
		case "nowiki":
			fmt.Fprintf(&b, "<pre>%s</pre>", html.EscapeString(n.Text))
		case "unknown_plugin":
			fmt.Fprintf(&b, `<div class="unsupported"><strong>Unsupported plugin: %s</strong><pre>%s</pre></div>`, html.EscapeString(n.Plugin), html.EscapeString(n.Raw))
		}
	}
	return b.String()
}

func PreviewHTML(p model.Page) string {
	var b strings.Builder
	for i := 0; i < len(p.Nodes); i++ {
		n := p.Nodes[i]
		switch n.Type {
		case "heading":
			fmt.Fprintf(&b, `<h%s id="%s">%s</h%s>`, n.Level, anchorID(n.Text), previewInline(n.Text, p.ID), n.Level)
		case "paragraph":
			fmt.Fprintf(&b, "<p>%s</p>", previewInline(n.Text, p.ID))
		case "bullet_item", "ordered_item":
			fmt.Fprint(&b, renderListHTML(p.Nodes, &i, true, p.ID))
		case "table_row":
			renderTableHTML(&b, p.Nodes, &i, true, p.ID)
		case "quote":
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>", previewInline(n.Text, p.ID))
		case "hr":
			b.WriteString("<hr>")
		case "code":
			fmt.Fprintf(&b, "<pre><code>%s</code></pre>", html.EscapeString(n.Text))
		case "info", "warning", "note":
			fmt.Fprintf(&b, `<aside class="hint %s"><strong>%s</strong><p>%s</p></aside>`, n.Type, html.EscapeString(strings.ToUpper(n.Type)), previewInline(n.Text, p.ID))
		case "wrap":
			renderWrapPreview(&b, n, p.ID)
		case "include":
			fmt.Fprintf(&b, `<div class="include">Include %s: %s</div>`, html.EscapeString(n.Meta["mode"]), html.EscapeString(n.Target))
		case "nowiki":
			fmt.Fprintf(&b, "<pre>%s</pre>", html.EscapeString(n.Text))
		case "unknown_plugin":
			fmt.Fprintf(&b, `<div class="unsupported"><strong>Unsupported plugin: %s</strong><pre>%s</pre></div>`, html.EscapeString(n.Plugin), html.EscapeString(n.Raw))
		}
	}
	return b.String()
}

func renderTableHTML(b *strings.Builder, nodes []model.Node, i *int, preview bool, current string) {
	start := *i
	for *i < len(nodes) && nodes[*i].Type == "table_row" {
		(*i)++
	}
	end := *i - 1
	if end < start {
		return
	}
	b.WriteString(`<table class="dokuwiki-table">`)
	headerEnd := start
	for headerEnd <= end && tableHeaderRow(nodes[headerEnd]) {
		headerEnd++
	}
	if headerEnd > start {
		b.WriteString("<thead>")
		for j := start; j < headerEnd; j++ {
			renderTableHTMLRow(b, nodes[j], preview, current)
		}
		b.WriteString("</thead>")
	}
	if headerEnd <= end {
		b.WriteString("<tbody>")
		for j := headerEnd; j <= end; j++ {
			renderTableHTMLRow(b, nodes[j], preview, current)
		}
		b.WriteString("</tbody>")
	}
	b.WriteString("</table>")
	*i = end
}

func renderTableHTMLRow(b *strings.Builder, n model.Node, preview bool, current string) {
	b.WriteString("<tr>")
	for _, c := range n.Children {
		tag := "td"
		if c.Type == "table_header" {
			tag = "th"
		}
		span := ""
		if c.Meta != nil && c.Meta["colspan"] != "" {
			span = ` colspan="` + html.EscapeString(c.Meta["colspan"]) + `"`
		}
		text := Inline(c.Text)
		if preview {
			text = previewInline(c.Text, current)
		}
		fmt.Fprintf(b, "<%s%s>%s</%s>", tag, span, text, tag)
	}
	b.WriteString("</tr>")
}

func tableHeaderRow(n model.Node) bool {
	if len(n.Children) == 0 {
		return false
	}
	for _, c := range n.Children {
		if c.Type != "table_header" {
			return false
		}
	}
	return true
}

func renderListHTML(nodes []model.Node, i *int, preview bool, current string) string {
	start := *i
	out := renderListLevelHTML(nodes, &start, nodeIndent(nodes[start]), preview, current)
	*i = start - 1
	return out
}

func renderListLevelHTML(nodes []model.Node, i *int, indent int, preview bool, current string) string {
	if *i >= len(nodes) || !isListNode(nodes[*i]) {
		return ""
	}
	typ := nodes[*i].Type
	tag := "ul"
	if typ == "ordered_item" {
		tag = "ol"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<%s>", tag)
	for *i < len(nodes) {
		n := nodes[*i]
		if !isListNode(n) {
			break
		}
		level := nodeIndent(n)
		if level < indent || level > indent || n.Type != typ {
			break
		}
		text := Inline(n.Text)
		if preview {
			text = previewInline(n.Text, current)
		}
		fmt.Fprintf(&b, "<li>%s", text)
		(*i)++
		for *i < len(nodes) && isListNode(nodes[*i]) && nodeIndent(nodes[*i]) > indent {
			nestedIndent := nodeIndent(nodes[*i])
			b.WriteString(renderListLevelHTML(nodes, i, nestedIndent, preview, current))
		}
		b.WriteString("</li>")
	}
	fmt.Fprintf(&b, "</%s>", tag)
	return b.String()
}

func isListNode(n model.Node) bool {
	return n.Type == "bullet_item" || n.Type == "ordered_item"
}

func renderWrapHTML(b *strings.Builder, n model.Node) { renderWrapBlock(b, n, false, "") }
func renderWrapPreview(b *strings.Builder, n model.Node, current string) { renderWrapBlock(b, n, true, current) }

func renderWrapBlock(b *strings.Builder, n model.Node, preview bool, current string) {
	classes, kind, width, align := "", "", "", ""
	if n.Meta != nil {
		classes, kind, width, align = n.Meta["classes"], n.Meta["kind"], n.Meta["width"], n.Meta["align"]
	}
	if kind == "" { kind = firstWrapClass(classes) }
	if align == "" { align = wrapClassAlignment(classes) }
	if kind == "clear" || hasWrapClass(classes, "clear") {
		b.WriteString(`<div class="dokuwiki-clear" style="clear:both;"></div>`)
		return
	}
	style := wrapLayoutStyle(kind, width, align)
	if style != "" {
		fmt.Fprintf(b, `<div class="dokuwiki-wrap %s" style="%s">`, html.EscapeString(classes), html.EscapeString(style))
	} else {
		fmt.Fprintf(b, `<div class="dokuwiki-wrap %s">`, html.EscapeString(classes))
	}
	for _, c := range n.Children {
		switch c.Type {
		case "paragraph":
			if preview { fmt.Fprintf(b, "<p>%s</p>", previewInline(c.Text, current)) } else { fmt.Fprintf(b, "<p>%s</p>", Inline(c.Text)) }
		case "heading":
			if preview { fmt.Fprintf(b, `<h%s id="%s">%s</h%s>`, c.Level, anchorID(c.Text), previewInline(c.Text, current), c.Level) } else { fmt.Fprintf(b, "<h%s>%s</h%s>", c.Level, Inline(c.Text), c.Level) }
		case "wrap":
			renderWrapBlock(b, c, preview, current)
		default:
			if c.Text != "" {
				if preview { fmt.Fprintf(b, "<p>%s</p>", previewInline(c.Text, current)) } else { fmt.Fprintf(b, "<p>%s</p>", Inline(c.Text)) }
			}
		}
	}
	b.WriteString("</div>")
}

func wrapLayoutStyle(kind, width, align string) string {
	var parts []string
	if kind == "group" { parts = append(parts, "display:flex;", "flex-wrap:wrap;", "width:100%;") }
	if kind == "column" && wrapWidth.MatchString(width) { parts = append(parts, "width:"+width+";") }
	switch strings.ToLower(strings.TrimSpace(align)) {
	case "left", "center", "right": parts = append(parts, "text-align:"+strings.ToLower(strings.TrimSpace(align))+";")
	}
	return strings.Join(parts, "")
}

func firstWrapClass(classes string) string { for _, c := range strings.Fields(classes) { return c }; return "" }
func hasWrapClass(classes, want string) bool { for _, c := range strings.Fields(classes) { if c == want { return true } }; return false }
func wrapClassAlignment(classes string) string { for _, c := range strings.Fields(classes) { switch strings.ToLower(c) { case "left", "center", "right": return strings.ToLower(c) } }; return "" }
