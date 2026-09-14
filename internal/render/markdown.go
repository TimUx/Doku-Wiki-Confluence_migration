package render

var lineBreak = regexp.MustCompile(`\\\\(\\s|$)`)

import (
	"regexp"
	"fmt"
	"strings"

	"github.com/TimUx/Doku-Wiki-Confluence_migration/internal/model"
)

func Markdown(p model.Page) string {
	var b strings.Builder
	for i := 0; i < len(p.Nodes); i++ {
		n := p.Nodes[i]
		switch n.Type {
		case "heading":
			fmt.Fprintf(&b, "%s %s\n\n", strings.Repeat("#", len(n.Level)), markdownInline(n.Text))
		case "bullet_item", "ordered_item":
			fmt.Fprint(&b, renderListMarkdown(p.Nodes, &i))
		case "table_row":
			renderTableMarkdown(&b, p.Nodes, &i)
		case "code":
			lang := ""
			if n.Meta != nil {
				lang = strings.TrimSpace(n.Meta["language"])
			}
			fmt.Fprintf(&b, "```%s\n%s\n```\n\n", lang, n.Text)
		case "include":
			fmt.Fprintf(&b, "> Include %s: `%s`\n\n", n.Meta["mode"], n.Target)
		default:
			fmt.Fprintf(&b, "%s\n\n", markdownInline(n.Text))
		}
	}
	return b.String()
}

func markdownInline(s string) string {
	s = sub.ReplaceAllString(s, "<sub>$1</sub>")
	s = sup.ReplaceAllString(s, "<sup>$1</sup>")
	s = lineBreak.ReplaceAllString(s, "<br/>")
	return s
}

func renderTableMarkdown(b *strings.Builder, nodes []model.Node, i *int) {
	start := *i
	for *i < len(nodes) && nodes[*i].Type == "table_row" { (*i)++ }
	end := *i - 1
	if end < start { return }
	if tableHeaderRow(nodes[start]) {
		writeTableMarkdownRow(b, nodes[start])
		b.WriteString("|")
		for range nodes[start].Children { b.WriteString(" --- |") }
		b.WriteByte('\n')
		for j := start + 1; j <= end; j++ { writeTableMarkdownRow(b, nodes[j]) }
	} else {
		for j := start; j <= end; j++ { writeTableMarkdownRow(b, nodes[j]) }
	}
	*i = end
}

func writeTableMarkdownRow(b *strings.Builder, n model.Node) {
	b.WriteString("| ")
	for j, c := range n.Children {
		if j > 0 { b.WriteString(" | ") }
		b.WriteString(strings.ReplaceAll(c.Text, "|", "\\|"))
	}
	b.WriteString(" |\n")
}

func renderListMarkdown(nodes []model.Node, i *int) string {
	start := *i
	out := renderListLevelMarkdown(nodes, &start, nodeIndent(nodes[start]), 0)
	*i = start - 1
	return out
}

func renderListLevelMarkdown(nodes []model.Node, i *int, indent, depth int) string {
	if *i >= len(nodes) || !isListNode(nodes[*i]) { return "" }
	var b strings.Builder
	typ := nodes[*i].Type
	counter := 1
	for *i < len(nodes) {
		n := nodes[*i]
		if !isListNode(n) || nodeIndent(n) < indent || nodeIndent(n) > indent || n.Type != typ { break }
		prefix := "- "
		if typ == "ordered_item" { prefix = fmt.Sprintf("%d. ", counter); counter++ }
		fmt.Fprintf(&b, "%s%s%s\n", strings.Repeat("  ", depth), prefix, markdownInline(n.Text))
		(*i)++
		for *i < len(nodes) && isListNode(nodes[*i]) && nodeIndent(nodes[*i]) > indent {
			nestedIndent := nodeIndent(nodes[*i])
			b.WriteString(renderListLevelMarkdown(nodes, i, nestedIndent, depth+1))
		}
	}
	return b.String()
}
